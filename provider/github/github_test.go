package github

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/ViBiOh/auth/auth"
	"github.com/ViBiOh/auth/provider"
	"github.com/ViBiOh/httputils"
	"golang.org/x/oauth2"
)

func Test_Flags(t *testing.T) {
	var cases = []struct {
		intention string
		want      int
	}{
		{
			`should return map with two entries`,
			2,
		},
	}

	for _, testCase := range cases {
		if result := Flags(``); len(result) != testCase.want {
			t.Errorf("%s\nFlags() = %+v, want %+v", testCase.intention, result, testCase.want)
		}
	}
}

func Test_Init(t *testing.T) {
	name := `GitHub`

	var cases = []struct {
		intention string
		config    map[string]interface{}
		want      bool
	}{
		{
			`should not initialize config if not client ID`,
			nil,
			false,
		},
		{
			`should init oauth config`,
			map[string]interface{}{`clientID`: &name, `clientSecret`: &name},
			true,
		},
	}

	for _, testCase := range cases {
		authClient := Auth{}
		authClient.Init(testCase.config)

		if result := authClient.oauthConf != nil; result != testCase.want {
			t.Errorf("%s\nInit(%+v) = %+v, want %+v", testCase.intention, testCase.config, authClient.oauthConf, testCase.want)
		}
	}
}

func Test_GetName(t *testing.T) {
	var cases = []struct {
		intention string
		want      string
	}{
		{
			`should return constant`,
			`GitHub`,
		},
	}

	for _, testCase := range cases {
		if result := (&Auth{}).GetName(); result != testCase.want {
			t.Errorf("%s\nGetName() = %+v, want %+v", testCase.intention, result, testCase.want)
		}
	}
}

func Test_GetUser(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(`Authorization`) == `token unauthorized` {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.Write([]byte(strings.TrimPrefix(r.Header.Get(`Authorization`), `token `)))
		}
	}))
	defer testServer.Close()

	var cases = []struct {
		intention string
		header    string
		want      *auth.User
		wantErr   error
	}{
		{
			`should handle fetching error`,
			`unauthorized`,
			nil,
			errors.New(`Error while fetching user informations: Error status 401`),
		},
		{
			`should handle malformed json`,
			`{"id":1,"login":"vibioh"`,
			nil,
			errors.New(`Error while unmarshalling user informations: unexpected end of JSON input`),
		},
		{
			`should handle valid request`,
			`{"id":1,"login":"vibioh"}`,
			&auth.User{ID: 1, Username: `vibioh`},
			nil,
		},
	}

	var failed bool

	for _, testCase := range cases {
		userURL = testServer.URL
		result, err := (&Auth{}).GetUser(testCase.header)

		failed = false

		if err == nil && testCase.wantErr != nil {
			failed = true
		} else if err != nil && testCase.wantErr == nil {
			failed = true
		} else if err != nil && err.Error() != testCase.wantErr.Error() {
			failed = true
		} else if !reflect.DeepEqual(result, testCase.want) {
			failed = true
		}

		if failed {
			t.Errorf("%s\nGetUser(%+v) = (%+v, %+v), want (%+v, %+v)", testCase.intention, testCase.header, result, err, testCase.want, testCase.wantErr)
		}
	}
}

func Test_GetAccessToken(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := httputils.ReadBody(r.Body)
		if strings.HasPrefix(string(body), `code=invalid`) {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.Write([]byte(`access_token=github_token`))
		}
	}))
	defer testServer.Close()

	endpoint = oauth2.Endpoint{
		AuthURL:  testServer.URL,
		TokenURL: testServer.URL,
	}

	configValue := `test`
	authClient := Auth{}
	authClient.Init(map[string]interface{}{
		`clientID`:     &configValue,
		`clientSecret`: &configValue,
	})

	var cases = []struct {
		intention string
		state     string
		code      string
		want      string
		wantErr   error
	}{
		{
			`should identify invalid state`,
			`state`,
			``,
			``,
			provider.ErrInvalidState,
		},
		{
			`should identify invalid code`,
			`test`,
			`invalidcode`,
			``,
			provider.ErrInvalidCode,
		},
		{
			`should return given token`,
			`test`,
			`validcode`,
			`github_token`,
			nil,
		},
	}

	var failed bool

	for _, testCase := range cases {
		result, err := authClient.GetAccessToken(configValue, testCase.state, testCase.code)

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
			t.Errorf("%s\nGetAccessToken(%+v, %+v) = (%+v, %+v), want (%+v, %+v)", testCase.intention, testCase.state, testCase.code, result, err, testCase.want, testCase.wantErr)
		}
	}
}
