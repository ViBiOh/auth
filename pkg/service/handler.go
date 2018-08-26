package service

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/auth/pkg/provider"
	"github.com/ViBiOh/auth/pkg/provider/basic"
	"github.com/ViBiOh/auth/pkg/provider/github"
	"github.com/ViBiOh/httputils/pkg/httperror"
	"github.com/ViBiOh/httputils/pkg/tools"
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

// NewBasicApp creates new App from Flags' config only for basic auth wrapper
func NewBasicApp(basicConfig map[string]interface{}) *App {
	return &App{
		providers: initProviders(map[string]providerConfig{
			`Basic`: {config: basicConfig, factory: basic.NewAuth},
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
func (a App) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			if _, err := w.Write(nil); err != nil {
				httperror.InternalServerError(w, err)
			}
			return
		}

		if r.Method != http.MethodGet {
			http.Error(w, fmt.Sprintf(`%d Method not allowed`, http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

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
	})
}