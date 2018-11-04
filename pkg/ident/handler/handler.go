package handler

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/auth/pkg/ident"
	"github.com/ViBiOh/httputils/pkg/httperror"
	"github.com/ViBiOh/httputils/pkg/tools"
)

const (
	loginPrefix    = `/login`
	redirectPrefix = `/redirect`
)

// App stores informations and secret of API
type App struct {
	providers    []ident.Auth
	redirect     string
	cookieDomain string
}

// NewApp creates new App from Flags' config
func NewApp(config map[string]*string, providers []ident.Auth) *App {
	return &App{
		redirect:     *config[`redirect`],
		cookieDomain: *config[`cookieDomain`],
		providers:    providers,
	}
}

// Flags add flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`redirect`:     flag.String(tools.ToCamel(fmt.Sprintf(`%sAuthRedirect`, prefix)), ``, `[auth] Redirect URL on Auth Success`),
		`cookieDomain`: flag.String(tools.ToCamel(fmt.Sprintf(`%sCookieDomain`, prefix)), ``, `[auth] Cookie Domain to Store Authentification`),
	}
}

// Handler for net/http package handling auth requests
func (a App) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodOptions:
			if _, err := w.Write(nil); err != nil {
				httperror.InternalServerError(w, err)
			}
			break

		case http.MethodGet:
			if r.URL.Path == `/user` {
				a.userHandler(w, r)
			} else if r.URL.Path == `/logout` {
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
			http.Error(w, fmt.Sprintf(`%d Method not allowed`, http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	})
}
