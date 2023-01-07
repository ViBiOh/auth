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
	t.Parallel()

	basicAuthRequest, _ := request.Get("/").BasicAuth("admin", "password").Build(context.TODO(), nil)

	cases := map[string]struct {
		instance   App
		request    *http.Request
		want       string
		wantStatus int
	}{
		"no provider": {
			New(nil, nil),
			httptest.NewRequest(http.MethodOptions, "/", nil),
			"OPTIONS",
			http.StatusOK,
		},
		"options": {
			New(nil, nil, testProvider{}),
			httptest.NewRequest(http.MethodOptions, "/", nil),
			"",
			http.StatusNoContent,
		},
		"failure": {
			New(nil, nil, testProvider{}),
			httptest.NewRequest(http.MethodGet, "/", nil),
			"empty authorization content\n",
			http.StatusTeapot,
		},
		"success": {
			New(nil, nil, testProvider{matching: true}),
			basicAuthRequest,
			"GET",
			http.StatusOK,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if _, err := w.Write([]byte(r.Method)); err != nil {
					t.Errorf("write: %s", err)
				}
			})

			writer := httptest.NewRecorder()
			testCase.instance.Middleware(handler).ServeHTTP(writer, testCase.request)

			if got := writer.Code; got != testCase.wantStatus {
				t.Errorf("Middleware = %d, want %d", got, testCase.wantStatus)
			}

			if got, _ := request.ReadBodyResponse(writer.Result()); string(got) != testCase.want {
				t.Errorf("Middleware = `%s`, want `%s`", string(got), testCase.want)
			}
		})
	}
}

func TestIsAuthenticated(t *testing.T) {
	t.Parallel()

	basicAuthRequest, _ := request.Get("/").BasicAuth("admin", "password").Build(context.TODO(), nil)
	errorRequest, _ := request.Get("/").Header("Authorization", "Basic").Build(context.TODO(), nil)

	cases := map[string]struct {
		instance App
		request  *http.Request
		want     model.User
		wantErr  error
	}{
		"no provider": {
			New(nil, nil),
			httptest.NewRequest(http.MethodGet, "/", nil),
			model.User{},
			ErrNoMatchingProvider,
		},
		"empty request": {
			New(testProvider{}, nil, testProvider{}),
			httptest.NewRequest(http.MethodGet, "/", nil),
			model.User{},
			ErrEmptyAuth,
		},
		"no match": {
			New(testProvider{}, nil, testProvider{}),
			basicAuthRequest,
			model.User{},
			ErrNoMatchingProvider,
		},
		"error on get user": {
			New(testProvider{}, nil, testProvider{matching: true}),
			errorRequest,
			model.User{},
			errTestProvider,
		},
		"valid": {
			New(testProvider{}, nil, testProvider{matching: true}),
			basicAuthRequest,
			model.NewUser(8000, "admin"),
			nil,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			_, got, gotErr := testCase.instance.IsAuthenticated(testCase.request)

			failed := false

			if testCase.wantErr == nil && gotErr != nil {
				failed = true
			} else if testCase.wantErr != nil && !errors.Is(gotErr, testCase.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, testCase.want) {
				failed = true
			}

			if failed {
				t.Errorf("IsAuthenticated() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestIsAuthorized(t *testing.T) {
	t.Parallel()

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
			New(nil, nil),
			args{
				context: model.StoreUser(context.TODO(), model.User{}),
				profile: "admin",
			},
			false,
		},
		"call provider": {
			New(testProvider{}, nil),
			args{
				context: model.StoreUser(context.TODO(), model.User{}),
				profile: "admin",
			},
			true,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if got := testCase.instance.IsAuthorized(testCase.args.context, testCase.args.profile); got != testCase.want {
				t.Errorf("IsAuthorized() = %t, want %t", got, testCase.want)
			}
		})
	}
}

func TestOnHandlerFail(t *testing.T) {
	t.Parallel()

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

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			writer := httptest.NewRecorder()
			onHandlerFail(writer, testCase.request, testCase.err, testCase.provider)

			if got := writer.Code; got != testCase.wantStatus {
				t.Errorf("onHandlerFail = %d, want %d", got, testCase.wantStatus)
			}

			if got, _ := request.ReadBodyResponse(writer.Result()); string(got) != testCase.want {
				t.Errorf("onHandlerFail = `%s`, want `%s`", string(got), testCase.want)
			}
		})
	}
}
