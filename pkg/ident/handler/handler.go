package handler

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/auth/pkg/ident"
	"github.com/ViBiOh/httputils/v2/pkg/errors"
	"github.com/ViBiOh/httputils/v2/pkg/httperror"
	"github.com/ViBiOh/httputils/v2/pkg/logger"
	"github.com/ViBiOh/httputils/v2/pkg/tools"
)

const (
	loginPrefix    = "/login"
	redirectPrefix = "/redirect"
)

// Config of package
type Config struct {
	cookieDomain *string
	redirect     *string
}

// App of package
type App struct {
	cookieDomain string
	providers    []ident.Auth
	redirect     string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		cookieDomain: tools.NewFlag(prefix, "auth").Name("CookieDomain").Default("").Label("Cookie Domain to Store Authentification").ToString(fs),
		redirect:     tools.NewFlag(prefix, "auth").Name("AuthRedirect").Default("").Label("Redirect URL on Auth Success").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config, providers []ident.Auth) *App {
	usedProviders := make([]ident.Auth, 0)
	for _, provider := range providers {
		if provider != nil {
			usedProviders = append(usedProviders, provider)
			logger.Info("Provider for %s", provider.GetName())
		}
	}

	return &App{
		redirect:     *config.redirect,
		cookieDomain: *config.cookieDomain,
		providers:    usedProviders,
	}
}

// Handler for net/http package handling auth requests
func (a App) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodOptions:
			if _, err := w.Write(nil); err != nil {
				httperror.InternalServerError(w, errors.WithStack(err))
				return
			}
			break

		case http.MethodGet:
			if r.URL.Path == "/user" {
				a.userHandler(w, r)
			} else if r.URL.Path == "/logout" {
				a.logoutHandler(w, r)
			} else if strings.HasPrefix(r.URL.Path, loginPrefix) {
				a.loginHandler(w, r)
			} else if strings.HasPrefix(r.URL.Path, redirectPrefix) {
				a.redirectHandler(w, r)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
			break

		default:
			http.Error(w, fmt.Sprintf("%d Method not allowed", http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	})
}
