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
	return model.NoneUser, errInvalidCredentials
}

func TestIsMatching(t *testing.T) {
	type args struct {
		content string
	}

	var cases = []struct {
		intention string
		args      args
		want      bool
	}{
		{
			"invalid",
			args{
				content: "c2VjcmV0Cg==",
			},
			false,
		},
		{
			"valid",
			args{
				content: "Basic c2VjcmV0Cg==",
			},
			true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := New(testProvider{}, "").IsMatching(tc.args.content); got != tc.want {
				t.Errorf("IsMatching() = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	type args struct {
		content string
	}

	var cases = []struct {
		intention string
		args      args
		want      model.User
		wantErr   error
	}{
		{
			"invalid base64",
			args{
				content: "🤪",
			},
			model.NoneUser,
			ident.ErrMalformedAuth,
		},
		{
			"invalid auth",
			args{
				content: "Basic c2VjcmV0Cg==",
			},
			model.NoneUser,
			ident.ErrMalformedAuth,
		},
		{
			"valid",
			args{
				content: "Basic YWRtaW46c2VjcmV0Cg==",
			},
			model.NewUser(1, "admin"),
			nil,
		},
		{
			"invalid",
			args{
				content: "Basic YWRtaW46YWRtaW4K",
			},
			model.NoneUser,
			errInvalidCredentials,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			got, gotErr := New(testProvider{}, "").GetUser(context.Background(), tc.args.content)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("GetUser() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}

func TestOnError(t *testing.T) {
	type args struct {
		realm string
		err   error
	}

	wantedHeader := http.Header{}
	wantedHeader.Add("WWW-Authenticate", "Basic charset=\"UTF-8\"")

	wantedRealmHeader := http.Header{}
	wantedRealmHeader.Add("WWW-Authenticate", "Basic realm=\"Testing\" charset=\"UTF-8\"")

	var cases = []struct {
		intention  string
		request    *http.Request
		args       args
		want       string
		wantStatus int
		wantHeader http.Header
	}{
		{
			"simple",
			httptest.NewRequest(http.MethodGet, "/", nil),
			args{
				err: errInvalidCredentials,
			},
			"invalid credentials\n",
			http.StatusUnauthorized,
			wantedHeader,
		},
		{
			"realm",
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

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			New(testProvider{}, tc.args.realm).OnError(writer, tc.request, tc.args.err)

			if got := writer.Code; got != tc.wantStatus {
				t.Errorf("OnError = %d, want %d", got, tc.wantStatus)
			}

			if got, _ := request.ReadBodyResponse(writer.Result()); string(got) != tc.want {
				t.Errorf("OnError = `%s`, want `%s`", string(got), tc.want)
			}

			for key := range tc.wantHeader {
				want := tc.wantHeader.Get(key)
				if got := writer.Header().Get(key); got != want {
					t.Errorf("`%s` Header = `%s`, want `%s`", key, got, want)
				}
			}
		})
	}
}
