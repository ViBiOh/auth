package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/ViBiOh/auth/pkg/auth"
	"github.com/ViBiOh/auth/pkg/ident"
	"github.com/ViBiOh/auth/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/httperror"
	httpmodel "github.com/ViBiOh/httputils/v3/pkg/model"
)

type key int

const (
	authorizationHeader     = "Authorization"
	ctxUserKey          key = iota
)

var _ httpmodel.Middleware = &app{}

var (
	// ErrEmptyAuth occurs when authorization content is not found
	ErrEmptyAuth = errors.New("empty authorization content")

	// ErrNoMatchingProvider occurs no provider is found for given auth
	ErrNoMatchingProvider = errors.New("no matching provider for Authrization content")
)

// App of package
type App interface {
	Handler(http.Handler) http.Handler
	IsAuthenticated(*http.Request) (ident.Provider, model.User, error)
}

type app struct {
	identProviders []ident.Provider
}

// New creates new App for given providers
func New(identProviders []ident.Provider) App {
	return &app{
		identProviders: identProviders,
	}
}

// UserFromContext retrieves user from context
func UserFromContext(ctx context.Context) model.User {
	rawUser := ctx.Value(ctxUserKey)
	if rawUser == nil {
		return model.NoneUser

	}

	if user, ok := rawUser.(model.User); ok {
		return user
	}
	return model.NoneUser

}

// Handler wrap next authenticated handler
func (a app) Handler(next http.Handler) http.Handler {
	if len(a.identProviders) == 0 {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		provider, user, err := a.IsAuthenticated(r)
		if err != nil {
			onHandlerFail(w, r, err, provider)
			return
		}

		if next != nil {
			userContext := context.WithValue(r.Context(), ctxUserKey, user)
			next.ServeHTTP(w, r.WithContext(userContext))
		}
	})
}

// IsAuthenticated check if request has correct headers for authentification
func (a app) IsAuthenticated(r *http.Request) (ident.Provider, model.User, error) {
	authContent := strings.TrimSpace(r.Header.Get(authorizationHeader))
	if len(strings.TrimSpace(authContent)) == 0 {
		return nil, model.NoneUser, ErrEmptyAuth
	}

	for _, provider := range a.identProviders {
		if !provider.IsMatching(authContent) {
			continue
		}

		user, err := provider.GetUser(authContent)
		return provider, user, err
	}

	return nil, model.NoneUser, ErrNoMatchingProvider
}

func onHandlerFail(w http.ResponseWriter, r *http.Request, err error, provider ident.Provider) {
	if err == ErrEmptyAuth {
		httperror.Unauthorized(w, err)
	} else if err == auth.ErrForbidden {
		httperror.Forbidden(w)
	} else {
		provider.OnError(w, r, err)
	}
}
