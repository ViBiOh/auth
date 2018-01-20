package service

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/auth/provider"
	"github.com/ViBiOh/auth/provider/basic"
	"github.com/ViBiOh/auth/provider/github"
	"github.com/ViBiOh/httputils/tools"
)

const (
	loginPrefix    = `/login`
	redirectPrefix = `/redirect`
)

// App stores informations and secret of API
type App struct {
	providers    []provider.Auth
	redirect     string
	cookieDomain string
}

// NewApp creates new App from Flags' config
func NewApp(config map[string]*string, basicConfig map[string]interface{}, githubConfig map[string]interface{}) *App {
	return &App{
		redirect:     *config[`redirect`],
		cookieDomain: *config[`cookieDomain`],
		providers: initProviders(map[string]providerConfig{
			`Basic`:  {config: basicConfig, factory: basic.NewAuth},
			`GitHub`: {config: githubConfig, factory: github.NewAuth},
		}),
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
func (a *App) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if r.URL.Path == `/user` {
			a.userHandler(w, r, a.providers)
		} else if r.URL.Path == `/logout` {
			a.logoutHandler(w, r)
		} else if strings.HasPrefix(r.URL.Path, loginPrefix) {
			a.loginHandler(w, r, a.providers)
		} else if strings.HasPrefix(r.URL.Path, redirectPrefix) {
			a.redirectHandler(w, r, a.providers)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})
}
