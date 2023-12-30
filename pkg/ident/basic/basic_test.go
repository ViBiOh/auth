package basic

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/ViBiOh/auth/v2/pkg/ident"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/request"
)

var errInvalidCredentials = errors.New("invalid credentials")

type testProvider struct{}

func (tp testProvider) Login(_ context.Context, login, password string) (model.User, error) {
	if login == "admin" && password == "secret" {
		return model.NewUser(1, "admin"), nil
	}
	return model.User{}, errInvalidCredentials
}

func TestIsMatching(t *testing.T) {
	t.Parallel()

	type args struct {
		content string
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"short": {
			args{
				content: "Bas",
			},
			false,
		},
		"invalid": {
			args{
				content: "c2VjcmV0Cg==",
			},
			false,
		},
		"valid": {
			args{
				content: "Basic c2VjcmV0Cg==",
			},
			true,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if got := New(testProvider{}, "").IsMatching(testCase.args.content); got != testCase.want {
				t.Errorf("IsMatching() = %t, want %t", got, testCase.want)
			}
		})
	}
}

func BenchmarkIsMatching(b *testing.B) {
	var service Service

	for i := 0; i < b.N; i++ {
		service.IsMatching("Basic abcdef1234567890")
	}
}

func TestGetUser(t *testing.T) {
	t.Parallel()

	type args struct {
		content string
	}

	cases := map[string]struct {
		args    args
		want    model.User
		wantErr error
	}{
		"invalid string": {
			args{
				content: "",
			},
			model.User{},
			ident.ErrMalformedAuth,
		},
		"invalid base64": {
			args{
				content: "Basic 🤪",
			},
			model.User{},
			ident.ErrMalformedAuth,
		},
		"invalid auth": {
			args{
				content: "Basic c2VjcmV0Cg==",
			},
			model.User{},
			ident.ErrMalformedAuth,
		},
		"valid": {
			args{
				content: "Basic YWRtaW46c2VjcmV0Cg==",
			},
			model.NewUser(1, "admin"),
			nil,
		},
		"invalid": {
			args{
				content: "Basic YWRtaW46YWRtaW4K",
			},
			model.User{},
			errInvalidCredentials,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			got, gotErr := New(testProvider{}, "").GetUser(context.Background(), testCase.args.content)

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
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			writer := httptest.NewRecorder()
			New(testProvider{}, testCase.args.realm).OnError(writer, testCase.request, testCase.args.err)

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
