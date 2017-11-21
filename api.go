package main

import (
	"errors"
	"flag"
	"log"
	"net/http"
	"strings"

	"github.com/ViBiOh/auth/cookie"

	"github.com/NYTimes/gziphandler"
	"github.com/ViBiOh/alcotest/alcotest"
	"github.com/ViBiOh/auth/auth"
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

const tokenPrefix = `/token`
const authorizePrefix = `/authorize`

var providers []provider.Auth
var errMalformedAuth = errors.New(`Malformed Authorization header`)

func initProvider(authProvider provider.Auth, config map[string]interface{}) provider.Auth {
	if err := authProvider.Init(config); err != nil {
		log.Fatalf(`Error while initializing %s auth: %v`, authProvider.GetName(), err)
	}

	return authProvider
}

// Init configures Auth providers
func Init(basicConfig map[string]interface{}, githubConfig map[string]interface{}) {
	providers = make([]provider.Auth, 2)

	providers[0] = initProvider(&basic.Auth{}, basicConfig)
	providers[1] = initProvider(&github.Auth{}, githubConfig)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get(`Authorization`)

	if authHeader == `` {
		httputils.Unauthorized(w, auth.ErrEmptyAuthorization)
		return
	}

	parts := strings.SplitN(authHeader, ` `, 2)
	if len(parts) != 2 {
		httputils.BadRequest(w, errMalformedAuth)
		return
	}

	for _, provider := range providers {
		if parts[0] == provider.GetName() {
			if user, err := provider.GetUser(parts[1]); err != nil {
				httputils.Unauthorized(w, err)
			} else {
				httputils.ResponseJSON(w, http.StatusOK, user, httputils.IsPretty(r.URL.RawQuery))
			}

			return
		}
	}

	httputils.BadRequest(w, provider.ErrUnknownAuthType)
}

func tokenHandler(w http.ResponseWriter, r *http.Request, oauthRedirect string) {
	for _, provider := range providers {
		if strings.HasSuffix(r.URL.Path, strings.ToLower(provider.GetName())) {
			cookieState, _ := cookie.GetCookieValue(r, `state`)

			if token, err := provider.GetAccessToken(cookieState, r.FormValue(`state`), r.FormValue(`code`)); err != nil {
				httputils.Unauthorized(w, err)
			} else if oauthRedirect != `` {
				http.SetCookie(w, &http.Cookie{
					Domain:   `vibioh.fr`,
					Name:     `auth`,
					MaxAge:   3600 * 24 * 7,
					Value:    `GitHub ` + token,
					Secure:   true,
					HttpOnly: true,
				})
				http.Redirect(w, r, oauthRedirect, http.StatusFound)
			} else {
				w.Write([]byte(token))
			}

			return
		}
	}

	httputils.BadRequest(w, provider.ErrUnknownTokenType)
}

func authorizeHandler(w http.ResponseWriter, r *http.Request, oauthRedirect string) {
	for _, provider := range providers {
		if strings.HasSuffix(r.URL.Path, strings.ToLower(provider.GetName())) {
			if redirect, headers, err := provider.Authorize(); err != nil {
				httputils.InternalServerError(w, err)
			} else {
				for key, value := range headers {
					w.Header().Add(key, value)
				}

				if redirect != `` {
					http.Redirect(w, r, redirect, http.StatusFound)
				} else {
					w.WriteHeader(http.StatusUnauthorized)
				}
			}

			return
		}
	}

	httputils.BadRequest(w, provider.ErrUnknownTokenType)
}

func handler(oauthRedirect string) http.Handler {
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
		} else if strings.HasPrefix(r.URL.Path, tokenPrefix) {
			tokenHandler(w, r, oauthRedirect)
		} else if strings.HasPrefix(r.URL.Path, authorizePrefix) {
			authorizeHandler(w, r, oauthRedirect)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})
}

func main() {
	port := flag.String(`port`, `1080`, `Listen port`)
	tls := flag.Bool(`tls`, true, `Serve TLS content`)
	oauthRedirect := flag.String(`redirect`, ``, `Redirect URI on OAuth Success`)
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

	Init(basicConfig, githubConfig)

	server := &http.Server{
		Addr:    `:` + *port,
		Handler: prometheus.Handler(prometheusConfig, rate.Handler(rateConfig, gziphandler.GzipHandler(owasp.Handler(owaspConfig, cors.Handler(corsConfig, handler(*oauthRedirect)))))),
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
