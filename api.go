package main

import (
	"errors"
	"flag"
	"log"
	"net/http"
	"strings"

	"github.com/NYTimes/gziphandler"
	"github.com/ViBiOh/alcotest/alcotest"
	"github.com/ViBiOh/auth/auth"
	"github.com/ViBiOh/auth/basic"
	"github.com/ViBiOh/auth/github"
	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/httputils/cert"
	"github.com/ViBiOh/httputils/cors"
	"github.com/ViBiOh/httputils/owasp"
	"github.com/ViBiOh/httputils/prometheus"
	"github.com/ViBiOh/httputils/rate"
)

const tokenPrefix = `/token`

// AuthProvider is a provider of Authentification methods
type AuthProvider interface {
	Init() error
	GetName() string
	GetUser(string) (*auth.User, error)
	GetAccessToken(requestState string, requestCode string) (string, error)
}

var providers = []AuthProvider{basic.Auth{}, github.Auth{}}

var errUnknownAuthType = errors.New(`Unknown authentication type`)
var errMalformedAuth = errors.New(`Malformed Authorization header`)
var errUnknownTokenType = errors.New(`Unknown token type`)

// Init configures Auth providers
func Init() {
	for _, provider := range providers {
		if err := provider.Init(); err != nil {
			log.Fatalf(`Error while initializing %s auth: %v`, provider.GetName(), err)
		}
	}
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

	httputils.BadRequest(w, errUnknownAuthType)
}

func tokenHandler(w http.ResponseWriter, r *http.Request) {
	for _, provider := range providers {
		if strings.HasSuffix(r.URL.Path, strings.ToLower(provider.GetName())) {
			if token, err := provider.GetAccessToken(r.FormValue(`state`), r.FormValue(`code`)); err != nil {
				httputils.Unauthorized(w, err)
			} else {
				w.Write([]byte(token))
			}

			return
		}
	}

	httputils.BadRequest(w, errUnknownTokenType)
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
		} else if strings.HasPrefix(r.URL.Path, tokenPrefix) {
			tokenHandler(w, r)
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
	flag.Parse()

	alcotest.DoAndExit(alcotestConfig)

	log.Printf(`Starting server on port %s`, *port)

	Init()

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
