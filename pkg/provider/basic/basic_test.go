package basic

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/request"
)

var errInvalidCredentials = errors.New("invalid credentials")

type testProvider struct{}

func (tp testProvider) GetBasicUser(_ context.Context, _ *http.Request, login, password string) (model.User, error) {
	if login == "admin" && password == "secret" {
		return model.NewUser(1, "admin"), nil
	}
	return model.User{}, errInvalidCredentials
}

func (tp testProvider) IsAuthorized(ctx context.Context, user model.User, profile string) bool {
	return true
}

func TestGetUser(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		request *http.Request
		want    model.User
		wantErr error
	}{
		"empty auth": {
			getRequestWithAuthorization(""),
			model.User{},
			model.ErrMalformedContent,
		},
		"invalid string": {
			getRequestWithAuthorization("c2VjcmV0Cg=="),
			model.User{},
			model.ErrMalformedContent,
		},
		"invalid base64": {
			getRequestWithAuthorization("Basic 🤪"),
			model.User{},
			model.ErrMalformedContent,
		},
		"invalid auth": {
			getRequestWithAuthorization("Basic c2VjcmV0Cg=="),
			model.User{},
			model.ErrMalformedContent,
		},
		"valid": {
			getRequestWithAuthorization("Basic YWRtaW46c2VjcmV0Cg=="),
			model.NewUser(1, "admin"),
			nil,
		},
		"invalid": {
			getRequestWithAuthorization("Basic YWRtaW46YWRtaW4K"),
			model.User{},
			errInvalidCredentials,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			got, gotErr := New(testProvider{}).GetUser(context.Background(), testCase.request)

			failed := false

			if testCase.wantErr == nil && gotErr != nil {
				failed = true
			} else if testCase.wantErr != nil && !errors.Is(gotErr, testCase.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, testCase.want) {
				failed = true
			}

			if failed {
				t.Errorf("GetUser() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestOnError(t *testing.T) {
	t.Parallel()

	type args struct {
		realm string
		err   error
	}

	wantedHeader := http.Header{}
	wantedHeader.Add("WWW-Authenticate", "Basic charset=\"UTF-8\"")

	wantedRealmHeader := http.Header{}
	wantedRealmHeader.Add("WWW-Authenticate", "Basic realm=\"Testing\" charset=\"UTF-8\"")

	cases := map[string]struct {
		request    *http.Request
		args       args
		want       string
		wantStatus int
		wantHeader http.Header
	}{
		"simple": {
			httptest.NewRequest(http.MethodGet, "/", nil),
			args{
				err: errInvalidCredentials,
			},
			"invalid credentials\n",
			http.StatusUnauthorized,
			wantedHeader,
		},
		"realm": {
			httptest.NewRequest(http.MethodGet, "/", nil),
			args{
				realm: "Testing",
				err:   errInvalidCredentials,
			},
			"invalid credentials\n",
			http.StatusUnauthorized,
			wantedRealmHeader,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			writer := httptest.NewRecorder()
			New(testProvider{}, WithRealm(testCase.args.realm)).OnError(writer, testCase.request, testCase.args.err)

			if got := writer.Code; got != testCase.wantStatus {
				t.Errorf("OnError = %d, want %d", got, testCase.wantStatus)
			}

			if got, _ := request.ReadBodyResponse(writer.Result()); string(got) != testCase.want {
				t.Errorf("OnError = `%s`, want `%s`", string(got), testCase.want)
			}

			for key := range testCase.wantHeader {
				want := testCase.wantHeader.Get(key)
				if got := writer.Header().Get(key); got != want {
					t.Errorf("`%s` Header = `%s`, want `%s`", key, got, want)
				}
			}
		})
	}
}

func BenchmarkGetUser(b *testing.B) {
	service := New(testProvider{})
	ctx := context.Background()

	req := getRequestWithAuthorization("Basic YWRtaW46c2VjcmV0Cg==")

	for b.Loop() {
		_, _ = service.GetUser(ctx, req)
	}
}

func getRequestWithAuthorization(auth string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	if len(auth) != 0 {
		req.Header.Add("Authorization", auth)
	}

	return req
}
