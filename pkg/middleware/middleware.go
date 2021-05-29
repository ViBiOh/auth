package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/ViBiOh/auth/v2/pkg/auth"
	"github.com/ViBiOh/auth/v2/pkg/ident"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/httperror"
	httpmodel "github.com/ViBiOh/httputils/v4/pkg/model"
)

var (
	_ httpmodel.Middleware = app{}.Middleware

	// ErrEmptyAuth occurs when authorization content is not found
	ErrEmptyAuth = errors.New("empty authorization content")

	// ErrNoMatchingProvider occurs no provider is found for given auth
	ErrNoMatchingProvider = errors.New("no matching provider for Authorization content")
)

// App of package
type App interface {
	Middleware(http.Handler) http.Handler
	IsAuthenticated(*http.Request, string) (ident.Provider, model.User, error)
	HasProfile(context.Context, model.User, string) bool
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

// Middleware wraps next authenticated handler
func (a app) Middleware(next http.Handler) http.Handler {
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
			next.ServeHTTP(w, r.WithContext(model.StoreUser(r.Context(), user)))
		}
	})
}

// IsAuthenticated check if request has correct headers for authentification
func (a app) IsAuthenticated(r *http.Request, profile string) (ident.Provider, model.User, error) {
	if len(a.identProviders) == 0 {
		return nil, model.NoneUser, ErrNoMatchingProvider
	}

	checkProfile := len(strings.TrimSpace(profile)) != 0

	authContent := strings.TrimSpace(r.Header.Get("Authorization"))
	if len(authContent) == 0 {
		return a.identProviders[0], model.NoneUser, ErrEmptyAuth
	}

	for _, provider := range a.identProviders {
		if !provider.IsMatching(authContent) {
			continue
		}

		user, err := provider.GetUser(r.Context(), authContent)
		if err != nil {
			return provider, user, err
		}

		if checkProfile && !a.HasProfile(r.Context(), user, profile) {
			return provider, user, auth.ErrForbidden
		}

		return provider, user, nil
	}

	return nil, model.NoneUser, ErrNoMatchingProvider
}

// HasProfile checks if User
func (a app) HasProfile(ctx context.Context, user model.User, profile string) bool {
	if a.authProvider == nil {
		return false
	}

	return a.authProvider.IsAuthorized(ctx, user, profile)
}

func onHandlerFail(w http.ResponseWriter, r *http.Request, err error, provider ident.Provider) {
	if err == auth.ErrForbidden {
		httperror.Forbidden(w)
	} else if provider != nil {
		provider.OnError(w, r, err)
	} else {
		httperror.BadRequest(w, err)
	}
}
