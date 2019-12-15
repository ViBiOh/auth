package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/ViBiOh/auth/v2/pkg/auth"
	"github.com/ViBiOh/auth/v2/pkg/ident"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/httperror"
	httpmodel "github.com/ViBiOh/httputils/v3/pkg/model"
)

type key int

const (
	ctxUserKey key = iota
)

var (
	_ httpmodel.Middleware = &app{}

	// ErrEmptyAuth occurs when authorization content is not found
	ErrEmptyAuth = errors.New("empty authorization content")

	// ErrNoMatchingProvider occurs no provider is found for given auth
	ErrNoMatchingProvider = errors.New("no matching provider for Authrization content")
)

// App of package
type App interface {
	Handler(http.Handler) http.Handler
	IsAuthenticated(*http.Request, string) (ident.Provider, model.User, error)
	HasProfile(model.User, string) bool
}

type app struct {
	authProvider   auth.Provider
	identProviders []ident.Provider
}

// New creates new App for given providers
func New(authProvider auth.Provider, identProviders ...ident.Provider) App {
	return &app{
		authProvider:   authProvider,
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

		provider, user, err := a.IsAuthenticated(r, "")
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
func (a app) IsAuthenticated(r *http.Request, profile string) (ident.Provider, model.User, error) {
	authContent := strings.TrimSpace(r.Header.Get("Authorization"))
	if len(strings.TrimSpace(authContent)) == 0 {
		return a.identProviders[0], model.NoneUser, ErrEmptyAuth
	}

	for _, provider := range a.identProviders {
		if !provider.IsMatching(authContent) {
			continue
		}

		user, err := provider.GetUser(authContent)
		if err != nil {
			return provider, user, err
		}

		if len(strings.TrimSpace(profile)) == 0 || a.HasProfile(user, profile) {
			return provider, user, nil
		}

		return provider, user, auth.ErrForbidden
	}

	return nil, model.NoneUser, ErrNoMatchingProvider
}

// HasProfile checks if User
func (a app) HasProfile(user model.User, profile string) bool {
	if a.authProvider == nil {
		return false
	}

	return a.authProvider.IsAuthorized(user, profile)
}

func onHandlerFail(w http.ResponseWriter, r *http.Request, err error, provider ident.Provider) {
	if err == auth.ErrForbidden {
		httperror.Forbidden(w)
	} else {
		provider.OnError(w, r, err)
	}
}
