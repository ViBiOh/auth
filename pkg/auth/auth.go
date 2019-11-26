package auth

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"strings"

	"github.com/ViBiOh/auth/pkg/ident"
	"github.com/ViBiOh/auth/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/flags"
	"github.com/ViBiOh/httputils/v3/pkg/httperror"
	httpmodel "github.com/ViBiOh/httputils/v3/pkg/model"
)

type key int

const (
	authorizationHeader     = "Authorization"
	ctxUserName         key = iota
)

var (
	// ErrForbidden occurs when user is identified but not authorized
	ErrForbidden = errors.New("forbidden access")

	_ httpmodel.Middleware = &app{}
)

// App of package
type App interface {
	Handler(http.Handler) http.Handler
	IsAuthenticated(*http.Request) (model.User, error)
}

// Config of package
type Config struct {
	disable *bool
}

type app struct {
	disabled bool
	provider ident.Provider
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		disable: flags.New(prefix, "auth").Name("Disable").Default(false).Label("Disable auth").ToBool(fs),
	}
}

// New creates new App from Config
func New(config Config) App {
	return &app{
		disabled: *config.disable,
	}
}

// NewService creates new App from Flags' config with service
func NewService(config Config, provider ident.Provider) App {
	return &app{
		disabled: *config.disable,
		provider: provider,
	}
}

// UserFromContext retrieves user from context
func UserFromContext(ctx context.Context) model.User {
	rawUser := ctx.Value(ctxUserName)
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
	if a.disabled {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		user, err := a.IsAuthenticated(r)
		if err != nil {
			a.onHandlerFail(w, r, err)
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxUserName, user)))
	})
}

// IsAuthenticated check if request has correct headers for authentification
func (a app) IsAuthenticated(r *http.Request) (model.User, error) {
	authContent := strings.TrimSpace(r.Header.Get(authorizationHeader))
	if len(strings.TrimSpace(authContent)) == 0 {
		return model.NoneUser, ident.ErrEmptyAuth
	}

	return a.provider.GetUser(r.Context(), authContent)
}

func (a app) onHandlerFail(w http.ResponseWriter, r *http.Request, err error) {
	if err == ErrForbidden {
		httperror.Forbidden(w)
	} else {
		a.provider.OnError(w, r, err)
	}
}
