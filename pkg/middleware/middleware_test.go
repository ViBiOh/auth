package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ViBiOh/auth/v3/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/request"
)

var errTestProvider = errors.New("decode")

type testProvider struct{}

func (t testProvider) GetUser(_ context.Context, _ http.ResponseWriter, r *http.Request) (model.User, error) {
	if r.Header.Get("Authorization") == "Basic Z3Vlc3Q6Z3Vlc3Q=" {
		return model.NewUser("guest"), nil
	} else if r.Header.Get("Authorization") == "Basic YWRtaW46cGFzc3dvcmQ=" {
		return model.NewUser("admin"), nil
	} else if r.Header.Get("Authorization") == "Basic" {
		return model.User{}, errTestProvider
	}

	return model.User{}, model.ErrMalformedContent
}

func (t testProvider) OnUnauthorized(w http.ResponseWriter, _ *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusTeapot)
}

func (t testProvider) IsAuthorized(_ context.Context, _ *http.Request, user model.User) bool {
	return user.Name == "admin"
}

func (t testProvider) OnForbidden(w http.ResponseWriter, _ *http.Request, user model.User) {
	http.Error(w, fmt.Sprintf("%s is not authorized", user.Name), http.StatusForbidden)
}

func TestMiddleware(t *testing.T) {
	t.Parallel()

	adminRequest, _ := request.Get("/").BasicAuth("admin", "password").Build(context.Background(), nil)
	guestRequest, _ := request.Get("/").BasicAuth("guest", "guest").Build(context.Background(), nil)

	cases := map[string]struct {
		instance   Service
		request    *http.Request
		want       string
		wantStatus int
	}{
		"options": {
			New(testProvider{}),
			httptest.NewRequest(http.MethodOptions, "/", nil),
			"",
			http.StatusNoContent,
		},
		"failure": {
			New(testProvider{}),
			httptest.NewRequest(http.MethodGet, "/", nil),
			model.ErrMalformedContent.Error() + "\n",
			http.StatusTeapot,
		},
		"forbidden": {
			New(testProvider{}, WithAuthorization(testProvider{})),
			guestRequest,
			"guest is not authorized\n",
			http.StatusForbidden,
		},
		"success": {
			New(testProvider{}),
			adminRequest,
			"GET",
			http.StatusOK,
		},
	}

	for intention, testCase := range cases {
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
