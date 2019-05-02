package github

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/ViBiOh/auth/pkg/ident"
	"github.com/ViBiOh/auth/pkg/model"
	"github.com/ViBiOh/httputils/pkg/cache"
	"github.com/ViBiOh/httputils/pkg/request"
	"golang.org/x/oauth2"
)

func TestNewAuth(t *testing.T) {
	empty := ""
	name := "GitHub"

	var cases = []struct {
		intention string
		config    Config
		want      bool
	}{
		{
			"should not initialize config if not client ID",
			Config{clientID: &empty, clientSecret: &empty, scopes: &empty},
			false,
		},
		{
			"should init oauth config",
			Config{clientID: &name, clientSecret: &name, scopes: &empty},
			true,
		},
	}

	for _, testCase := range cases {
		auth, _ := New(testCase.config)
		var authClient *App
		if auth != nil {
			authClient = auth.(*App)
		}

		if authClient != nil {
			if result := authClient.oauthConf != nil; result != testCase.want {
				t.Errorf("%s\nNewAuth(%+v) = %+v, want %+v", testCase.intention, testCase.config, authClient.oauthConf, testCase.want)
			}
		} else if testCase.want {
			t.Errorf("%s\nNewAuth(%+v) = %+v, want %+v", testCase.intention, testCase.config, nil, testCase.want)
		}
	}
}

func TestGetName(t *testing.T) {
	var cases = []struct {
		intention string
		want      string
	}{
		{
			"should return constant",
			"GitHub",
		},
	}

	for _, testCase := range cases {
		if result := (&App{}).GetName(); result != testCase.want {
			t.Errorf("%s\nGetName() = %+v, want %+v", testCase.intention, result, testCase.want)
		}
	}
}

func TestGetUser(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "token unauthorized" {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.Write([]byte(strings.TrimPrefix(r.Header.Get("Authorization"), "token ")))
		}
	}))
	defer testServer.Close()

	var cases = []struct {
		intention string
		header    string
		want      *model.User
		wantErr   error
	}{
		{
			"should handle fetching error",
			"unauthorized",
			nil,
			errors.New("error status 401"),
		},
		{
			"should handle malformed json",
			`{"id":1,"login":"vibioh"`,
			nil,
			errors.New("unexpected end of JSON input"),
		},
		{
			"should handle valid request",
			`{"id":1,"login":"vibioh"}`,
			&model.User{ID: "1", Username: "vibioh"},
			nil,
		},
	}

	var failed bool

	for _, testCase := range cases {
		userURL = testServer.URL
		result, err := (&App{
			oauthConf:  &oauth2.Config{},
			usersCache: cache.New(),
		}).GetUser(nil, testCase.header)

		failed = false

		if err == nil && testCase.wantErr != nil {
			failed = true
		} else if err != nil && testCase.wantErr == nil {
			failed = true
		} else if err != nil && !strings.Contains(err.Error(), testCase.wantErr.Error()) {
			failed = true
		} else if !reflect.DeepEqual(result, testCase.want) {
			failed = true
		}

		if failed {
			t.Errorf("%s\nGetUser(%+v) = (%+v, %+v), want (%+v, %+v)", testCase.intention, testCase.header, result, err, testCase.want, testCase.wantErr)
		}
	}
}

func TestLogin(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := request.ReadBodyRequest(r)
		if strings.Contains(string(body), "code=invalidcode") {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.Write([]byte("access_token=github_token"))
		}
	}))
	defer testServer.Close()

	endpoint = oauth2.Endpoint{
		AuthURL:  testServer.URL,
		TokenURL: testServer.URL,
	}

	configValue := "test"
	auth, _ := New(Config{
		clientID:     &configValue,
		clientSecret: &configValue,
		scopes:       &configValue,
	})
	authClient := auth.(*App)

	var cases = []struct {
		intention string
		request   *http.Request
		want      string
		wantErr   error
	}{
		{
			"should identify invalid state",
			httptest.NewRequest(http.MethodGet, "/?state=state", nil),
			"",
			ident.ErrInvalidState,
		},
		{
			"should identify invalid code",
			httptest.NewRequest(http.MethodGet, "/?state=test&code=invalidcode", nil),
			"",
			ident.ErrInvalidCode,
		},
		{
			"should return given token",
			httptest.NewRequest(http.MethodGet, "/?state=test&code=validcode", nil),
			"github_token",
			nil,
		},
	}

	var failed bool

	for _, testCase := range cases {
		authClient.states.Store(configValue, true)
		result, err := authClient.Login(testCase.request)

		failed = false

		if err == nil && testCase.wantErr != nil {
			failed = true
		} else if err != nil && testCase.wantErr == nil {
			failed = true
		} else if err != nil && err.Error() != testCase.wantErr.Error() {
			failed = true
		} else if result != testCase.want {
			failed = true
		}

		if failed {
			t.Errorf("%s\nLogin(%+v) = (%+v, %+v), want (%+v, %+v)", testCase.intention, testCase.request, result, err, testCase.want, testCase.wantErr)
		}
	}
}
