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
	"github.com/ViBiOh/httputils/v4/pkg/tracer"
)

var errTestProvider = errors.New("decode")

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
		return model.User{}, errTestProvider
	}
	return model.User{}, nil
}

func (t testProvider) OnError(w http.ResponseWriter, _ *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusTeapot)
}

func TestMiddleware(t *testing.T) {
	basicAuthRequest, _ := request.Get("/").BasicAuth("admin", "password").Build(context.Background(), nil)

	cases := map[string]struct {
		instance   App
		request    *http.Request
		want       string
		wantStatus int
	}{
		"no provider": {
			New(nil, tracer.App{}),
			httptest.NewRequest(http.MethodOptions, "/", nil),
			"OPTIONS",
			http.StatusOK,
		},
		"options": {
			New(nil, tracer.App{}, testProvider{}),
			httptest.NewRequest(http.MethodOptions, "/", nil),
			"",
			http.StatusNoContent,
		},
		"failure": {
			New(nil, tracer.App{}, testProvider{}),
			httptest.NewRequest(http.MethodGet, "/", nil),
			"empty authorization content\n",
			http.StatusTeapot,
		},
		"success": {
			New(nil, tracer.App{}, testProvider{matching: true}),
			basicAuthRequest,
			"GET",
			http.StatusOK,
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if _, err := w.Write([]byte(r.Method)); err != nil {
					t.Errorf("write: %s", err)
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
	basicAuthRequest, _ := request.Get("/").BasicAuth("admin", "password").Build(context.Background(), nil)
	errorRequest, _ := request.Get("/").Header("Authorization", "Basic").Build(context.Background(), nil)

	cases := map[string]struct {
		instance App
		request  *http.Request
		want     model.User
		wantErr  error
	}{
		"no provider": {
			New(nil, tracer.App{}),
			httptest.NewRequest(http.MethodGet, "/", nil),
			model.User{},
			ErrNoMatchingProvider,
		},
		"empty request": {
			New(testProvider{}, tracer.App{}, testProvider{}),
			httptest.NewRequest(http.MethodGet, "/", nil),
			model.User{},
			ErrEmptyAuth,
		},
		"no match": {
			New(testProvider{}, tracer.App{}, testProvider{}),
			basicAuthRequest,
			model.User{},
			ErrNoMatchingProvider,
		},
		"error on get user": {
			New(testProvider{}, tracer.App{}, testProvider{matching: true}),
			errorRequest,
			model.User{},
			errTestProvider,
		},
		"valid": {
			New(testProvider{}, tracer.App{}, testProvider{matching: true}),
			basicAuthRequest,
			model.NewUser(8000, "admin"),
			nil,
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			_, got, gotErr := tc.instance.IsAuthenticated(tc.request)

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

func TestIsAuthorized(t *testing.T) {
	type args struct {
		context context.Context
		profile string
	}

	cases := map[string]struct {
		instance App
		args     args
		want     bool
	}{
		"no provider": {
			New(nil, tracer.App{}),
			args{
				context: model.StoreUser(context.Background(), model.User{}),
				profile: "admin",
			},
			false,
		},
		"call provider": {
			New(testProvider{}, tracer.App{}),
			args{
				context: model.StoreUser(context.Background(), model.User{}),
				profile: "admin",
			},
			true,
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			if got := tc.instance.IsAuthorized(tc.args.context, tc.args.profile); got != tc.want {
				t.Errorf("IsAuthorized() = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestOnHandlerFail(t *testing.T) {
	cases := map[string]struct {
		request    *http.Request
		err        error
		provider   ident.Provider
		want       string
		wantStatus int
	}{
		"forbidden": {
			httptest.NewRequest(http.MethodGet, "/", nil),
			auth.ErrForbidden,
			nil,
			"⛔️\n",
			http.StatusForbidden,
		},
		"onError": {
			httptest.NewRequest(http.MethodOptions, "/", nil),
			ErrNoMatchingProvider,
			testProvider{},
			"no matching provider for Authorization content\n",
			http.StatusTeapot,
		},
		"no provider": {
			httptest.NewRequest(http.MethodOptions, "/", nil),
			ErrNoMatchingProvider,
			nil,
			"no matching provider for Authorization content\n",
			http.StatusBadRequest,
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
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
