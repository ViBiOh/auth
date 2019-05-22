package handler

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/auth/pkg/ident"
	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/httputils/pkg/httperror"
	"github.com/ViBiOh/httputils/pkg/logger"
	"github.com/ViBiOh/httputils/pkg/tools"
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
		cookieDomain: fs.String(tools.ToCamel(fmt.Sprintf("%sCookieDomain", prefix)), "", "[auth] Cookie Domain to Store Authentification"),
		redirect:     fs.String(tools.ToCamel(fmt.Sprintf("%sAuthRedirect", prefix)), "", "[auth] Redirect URL on Auth Success"),
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
