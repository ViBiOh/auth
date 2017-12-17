package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/NYTimes/gziphandler"
	"github.com/ViBiOh/alcotest/alcotest"
	"github.com/ViBiOh/auth/auth"
	"github.com/ViBiOh/auth/cookie"
	"github.com/ViBiOh/auth/provider"
	"github.com/ViBiOh/auth/provider/basic"
	"github.com/ViBiOh/auth/provider/github"
	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/httputils/cert"
	"github.com/ViBiOh/httputils/cors"
	"github.com/ViBiOh/httputils/owasp"
	"github.com/ViBiOh/httputils/prometheus"
	"github.com/ViBiOh/httputils/rate"
)

const loginPrefix = `/login`
const redirectPrefix = `/redirect`

var availableProviders = []provider.Auth{
	&basic.Auth{},
	&github.Auth{},
}

var providers []provider.Auth
var errMalformedAuth = errors.New(`Malformed Authorization content`)

var (
	authRedirect = flag.String(`authRedirect`, ``, `Redirect URL on Auth Success`)
	cookieDomain = flag.String(`cookieDomain`, ``, `Cookie Domain to Store Authentification`)
)

func initProvider(authProvider provider.Auth, config map[string]interface{}) bool {
	if err := authProvider.Init(config); err != nil {
		log.Printf(`Error while initializing %s provider: %v`, authProvider.GetName(), err)
		return false
	}

	return true
}

// Init configures Auth providers
func Init(providerConfig map[string]map[string]interface{}) {
	providers = make([]provider.Auth, 0, len(availableProviders))

	for _, provider := range availableProviders {
		if initProvider(provider, providerConfig[provider.GetName()]) {
			providers = append(providers, provider)
		}
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func getUser(r *http.Request) (*auth.User, error) {
	authContent := r.Header.Get(`Authorization`)

	if authContent == `` {
		return nil, auth.ErrEmptyAuthorization
	}

	parts := strings.SplitN(authContent, ` `, 2)
	if len(parts) != 2 {
		return nil, errMalformedAuth
	}

	for _, provider := range providers {
		if parts[0] == provider.GetName() {
			user, err := provider.GetUser(parts[1])
			if err != nil {
				return nil, err
			}
			return user, nil
		}
	}

	return nil, provider.ErrUnknownAuthType
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	user, err := getUser(r)
	if err != nil {
		if err == errMalformedAuth || err == provider.ErrUnknownAuthType {
			httputils.BadRequest(w, err)
		} else {
			httputils.Unauthorized(w, err)
		}

		return
	}

	httputils.ResponseJSON(w, http.StatusOK, user, httputils.IsPretty(r.URL.RawQuery))
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie.ClearCookieAndRedirect(w, r, *authRedirect, *cookieDomain)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	for _, provider := range providers {
		if strings.HasSuffix(r.URL.Path, strings.ToLower(provider.GetName())) {
			if redirect, err := provider.Redirect(); err != nil {
				httputils.InternalServerError(w, err)
			} else {
				http.Redirect(w, r, redirect, http.StatusFound)
			}

			return
		}
	}

	httputils.BadRequest(w, provider.ErrUnknownAuthType)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	for _, provider := range providers {
		if strings.HasSuffix(r.URL.Path, strings.ToLower(provider.GetName())) {
			if token, err := provider.Login(r); err != nil {
				w.Header().Add(`WWW-Authenticate`, provider.GetName())
				httputils.Unauthorized(w, err)
			} else if *authRedirect != `` {
				cookie.SetCookieAndRedirect(w, r, *authRedirect, *cookieDomain, fmt.Sprintf(`%s %s`, provider.GetName(), token))
			} else {
				w.Write([]byte(token))
			}

			return
		}
	}

	httputils.BadRequest(w, provider.ErrUnknownAuthType)
}

func handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Write(nil)
			return
		}

		if r.Method == http.MethodGet && r.URL.Path == `/health` {
			healthHandler(w, r)
			return
		}

		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if r.URL.Path == `/user` {
			userHandler(w, r)
		} else if r.URL.Path == `/logout` {
			logoutHandler(w, r)
		} else if strings.HasPrefix(r.URL.Path, loginPrefix) {
			loginHandler(w, r)
		} else if strings.HasPrefix(r.URL.Path, redirectPrefix) {
			redirectHandler(w, r)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})
}

func main() {
	port := flag.String(`port`, `1080`, `Listen port`)
	tls := flag.Bool(`tls`, true, `Serve TLS content`)
	alcotestConfig := alcotest.Flags(``)
	certConfig := cert.Flags(`tls`)
	prometheusConfig := prometheus.Flags(`prometheus`)
	rateConfig := rate.Flags(`rate`)
	owaspConfig := owasp.Flags(``)
	corsConfig := cors.Flags(`cors`)

	basicConfig := basic.Flags(`basic`)
	githubConfig := github.Flags(`github`)

	flag.Parse()

	alcotest.DoAndExit(alcotestConfig)

	log.Printf(`Starting server on port %s`, *port)

	Init(map[string]map[string]interface{}{
		`Basic`:  basicConfig,
		`GitHub`: githubConfig,
	})

	server := &http.Server{
		Addr:    `:` + *port,
		Handler: prometheus.Handler(prometheusConfig, rate.Handler(rateConfig, gziphandler.GzipHandler(owasp.Handler(owaspConfig, cors.Handler(corsConfig, handler()))))),
	}

	var serveError = make(chan error)
	go func() {
		defer close(serveError)
		if *tls {
			log.Print(`Listening with TLS enabled`)
			serveError <- cert.ListenAndServeTLS(certConfig, server)
		} else {
			log.Print(`⚠ auth is running without secure connection ⚠`)
			serveError <- server.ListenAndServe()
		}
	}()

	httputils.ServerGracefulClose(server, serveError, nil)
}
