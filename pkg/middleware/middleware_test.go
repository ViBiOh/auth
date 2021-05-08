package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/ViBiOh/auth/v2/pkg/auth"
	"github.com/ViBiOh/auth/v2/pkg/ident"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/request"
)

var errTestProvider = errors.New("unable to decode")

type testProvider struct {
	matching bool
}

func (t testProvider) IsAuthorized(_ context.Context, _ model.User, profile string) bool {
	return profile == "admin"
}

func (t testProvider) IsMatching(_ string) bool {
	return t.matching
}

func (t testProvider) GetUser(_ context.Context, input string) (model.User, error) {
	if input == "Basic YWRtaW46cGFzc3dvcmQ=" {
		return model.NewUser(8000, "admin"), nil
	} else if input == "Basic" {
		return model.NoneUser, errTestProvider
	}
	return model.NoneUser, nil
}

func (t testProvider) OnError(w http.ResponseWriter, _ *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusTeapot)
}

func TestMiddleware(t *testing.T) {
	basicAuthRequest, _ := request.New().BasicAuth("admin", "password").Get("/").Build(context.Background(), nil)

	var cases = []struct {
		intention  string
		instance   App
		request    *http.Request
		want       string
		wantStatus int
	}{
		{
			"no provider",
			New(nil),
			httptest.NewRequest(http.MethodOptions, "/", nil),
			"OPTIONS",
			http.StatusOK,
		},
		{
			"options",
			New(nil, testProvider{}),
			httptest.NewRequest(http.MethodOptions, "/", nil),
			"",
			http.StatusNoContent,
		},
		{
			"failure",
			New(nil, testProvider{}),
			httptest.NewRequest(http.MethodGet, "/", nil),
			"empty authorization content\n",
			http.StatusTeapot,
		},
		{
			"success",
			New(nil, testProvider{matching: true}),
			basicAuthRequest,
			"GET",
			http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if _, err := w.Write([]byte(r.Method)); err != nil {
					t.Errorf("unable to write: %s", err)
				}
			})

			writer := httptest.NewRecorder()
			tc.instance.Middleware(handler).ServeHTTP(writer, tc.request)

			if got := writer.Code; got != tc.wantStatus {
				t.Errorf("Middleware = %d, want %d", got, tc.wantStatus)
			}

			if got, _ := request.ReadBodyResponse(writer.Result()); string(got) != tc.want {
				t.Errorf("Middleware = `%s`, want `%s`", string(got), tc.want)
			}
		})
	}
}

func TestIsAuthenticated(t *testing.T) {
	basicAuthRequest, _ := request.New().BasicAuth("admin", "password").Get("/").Build(context.Background(), nil)
	errorRequest, _ := request.New().Header("Authorization", "Basic").Get("/").Build(context.Background(), nil)

	var cases = []struct {
		intention string
		instance  App
		request   *http.Request
		profile   string
		want      model.User
		wantErr   error
	}{
		{
			"no provider",
			New(nil),
			httptest.NewRequest(http.MethodGet, "/", nil),
			"",
			model.NoneUser,
			ErrNoMatchingProvider,
		},
		{
			"empty request",
			New(testProvider{}, testProvider{}),
			httptest.NewRequest(http.MethodGet, "/", nil),
			"",
			model.NoneUser,
			ErrEmptyAuth,
		},
		{
			"no match",
			New(testProvider{}, testProvider{}),
			basicAuthRequest,
			"",
			model.NoneUser,
			ErrNoMatchingProvider,
		},
		{
			"error on get user",
			New(testProvider{}, testProvider{matching: true}),
			errorRequest,
			"",
			model.NoneUser,
			errTestProvider,
		},
		{
			"no profile",
			New(testProvider{}, testProvider{matching: true}),
			basicAuthRequest,
			"",
			model.NewUser(8000, "admin"),
			nil,
		},
		{
			"admin profile",
			New(testProvider{}, testProvider{matching: true}),
			basicAuthRequest,
			"admin",
			model.NewUser(8000, "admin"),
			nil,
		},
		{
			"invalid profile",
			New(testProvider{}, testProvider{matching: true}),
			basicAuthRequest,
			"guest",
			model.NewUser(8000, "admin"),
			auth.ErrForbidden,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			_, got, gotErr := tc.instance.IsAuthenticated(tc.request, tc.profile)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("IsAuthenticated() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}

func TestHasProfile(t *testing.T) {
	var cases = []struct {
		intention string
		instance  App
		user      model.User
		profile   string
		want      bool
	}{
		{
			"no provider",
			New(nil),
			model.NoneUser,
			"admin",
			false,
		},
		{
			"call provider",
			New(testProvider{}),
			model.NoneUser,
			"admin",
			true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := tc.instance.HasProfile(context.Background(), tc.user, tc.profile); got != tc.want {
				t.Errorf("HasProfile() = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestOnHandlerFail(t *testing.T) {
	var cases = []struct {
		intention  string
		request    *http.Request
		err        error
		provider   ident.Provider
		want       string
		wantStatus int
	}{
		{
			"forbidden",
			httptest.NewRequest(http.MethodGet, "/", nil),
			auth.ErrForbidden,
			nil,
			"⛔️\n",
			http.StatusForbidden,
		},
		{
			"onError",
			httptest.NewRequest(http.MethodOptions, "/", nil),
			ErrNoMatchingProvider,
			testProvider{},
			"no matching provider for Authorization content\n",
			http.StatusTeapot,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			onHandlerFail(writer, tc.request, tc.err, tc.provider)

			if got := writer.Code; got != tc.wantStatus {
				t.Errorf("onHandlerFail = %d, want %d", got, tc.wantStatus)
			}

			if got, _ := request.ReadBodyResponse(writer.Result()); string(got) != tc.want {
				t.Errorf("onHandlerFail = `%s`, want `%s`", string(got), tc.want)
			}
		})
	}
}
