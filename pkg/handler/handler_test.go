package handler

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
	"github.com/ViBiOh/httputils/v3/pkg/request"
)

var errTestProvider = errors.New("unable to decode")

type testProvider struct {
	matching bool
}

func (t testProvider) IsAuthorized(user model.User, profile string) bool {
	if profile == "admin" {
		return true
	}
	return false
}

func (t testProvider) IsMatching(input string) bool {
	return t.matching
}

func (t testProvider) GetUser(input string) (model.User, error) {
	if input == "Basic YWRtaW46cGFzc3dvcmQ=" {
		return model.NewUser(8000, "admin"), nil
	} else if input == "Basic" {
		return model.NoneUser, errTestProvider
	}
	return model.NoneUser, nil
}

func (t testProvider) OnError(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusTeapot)
}

func TestUserFromContext(t *testing.T) {
	var cases = []struct {
		intention string
		input     context.Context
		want      model.User
	}{
		{
			"empty",
			context.Background(),
			model.NoneUser,
		},
		{
			"invalid type",
			context.WithValue(context.Background(), ctxUserKey, "User value"),
			model.NoneUser,
		},
		{
			"valid",
			context.WithValue(context.Background(), ctxUserKey, model.NewUser(8000, "test")),
			model.NewUser(8000, "test"),
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			if result := UserFromContext(testCase.input); !reflect.DeepEqual(result, testCase.want) {
				t.Errorf("UserFromContext() = %v, want %v", result, testCase.want)
			}
		})
	}
}

func TestHandler(t *testing.T) {
	basicAuthRequest, _ := request.New().BasicAuth("admin", "password").Get("/").Build(context.Background(), nil)

	var cases = []struct {
		intention  string
		instance   App
		request    *http.Request
		want       string
		wantStatus int
	}{
		{
			"do nothing if no provider",
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
			"handle authentication failure",
			New(nil, testProvider{}),
			httptest.NewRequest(http.MethodGet, "/", nil),
			"empty authorization content\n",
			http.StatusTeapot,
		},
		{
			"handle authentication",
			New(nil, testProvider{matching: true}),
			basicAuthRequest,
			"GET",
			http.StatusOK,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(r.Method))
			})

			writer := httptest.NewRecorder()
			testCase.instance.Handler(handler).ServeHTTP(writer, testCase.request)

			if result := writer.Code; result != testCase.wantStatus {
				t.Errorf("Handler = %d, want %d", result, testCase.wantStatus)
			}

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
				t.Errorf("Handler = `%s`, want `%s`", string(result), testCase.want)
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
			"no provider",
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

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			_, result, err := testCase.instance.IsAuthenticated(testCase.request, testCase.profile)

			failed := false

			if testCase.wantErr != nil && !errors.Is(err, testCase.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(result, testCase.want) {
				failed = true
			}

			if failed {
				t.Errorf("IsAuthenticated() = (%v, `%s`), want (%v, `%s`)", result, err, testCase.want, testCase.wantErr)
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

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			if result := testCase.instance.HasProfile(testCase.user, testCase.profile); result != testCase.want {
				t.Errorf("HasProfile() = %t, want %t", result, testCase.want)
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

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			onHandlerFail(writer, testCase.request, testCase.err, testCase.provider)

			if result := writer.Code; result != testCase.wantStatus {
				t.Errorf("onHandlerFail = %d, want %d", result, testCase.wantStatus)
			}

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
				t.Errorf("onHandlerFail = `%s`, want `%s`", string(result), testCase.want)
			}
		})
	}
}
