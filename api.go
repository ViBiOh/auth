package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/NYTimes/gziphandler"
	"github.com/ViBiOh/alcotest/alcotest"
	"github.com/ViBiOh/auth/basic"
	"github.com/ViBiOh/auth/github"
	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/httputils/cert"
	"github.com/ViBiOh/httputils/cors"
	"github.com/ViBiOh/httputils/owasp"
	"github.com/ViBiOh/httputils/prometheus"
	"github.com/ViBiOh/httputils/rate"
)

const basicPrefix = `Basic `
const githubPrefix = `GitHub `

// Init configures Auth providers
func Init() {
	if err := basic.Init(); err != nil {
		log.Fatalf(`Error while initializing Basic auth: %v`, err)
	}
	if err := github.Init(); err != nil {
		log.Fatalf(`Error while initializing GitHub auth: %v`, err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get(`Authorization`)

	if strings.HasPrefix(authHeader, basicPrefix) {
		if username, err := basic.GetUsername(strings.TrimPrefix(authHeader, basicPrefix)); err != nil {
			httputils.Unauthorized(w, err)
		} else {
			w.Write([]byte(username))
		}
	} else if strings.HasPrefix(authHeader, githubPrefix) {
		if username, err := github.GetUsername(strings.TrimPrefix(authHeader, githubPrefix)); err != nil {
			httputils.Unauthorized(w, err)
		} else {
			w.Write([]byte(username))
		}
	} else {
		httputils.BadRequest(w, fmt.Errorf(`Unable to read authentication type`))
	}
}

func githubTokenHandler(w http.ResponseWriter, r *http.Request) {
	if token, err := github.GetAccessToken(r.FormValue(`state`), r.FormValue(`code`)); err != nil {
		httputils.Unauthorized(w, err)
	} else {
		w.Write([]byte(token))
	}
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
		} else if r.URL.Path == `/token/github` {
			githubTokenHandler(w, r)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})
}

func main() {
	url := flag.String(`c`, ``, `URL to healthcheck (check and exit)`)
	port := flag.String(`port`, `1080`, `Listen port`)
	tls := flag.Bool(`tls`, true, `Serve TLS content`)
	prometheusConfig := prometheus.Flags(`prometheus`)
	rateConfig := rate.Flags(`rate`)
	owaspConfig := owasp.Flags(``)
	corsConfig := cors.Flags(`cors`)
	flag.Parse()

	if *url != `` {
		alcotest.Do(url)
		return
	}

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
			serveError <- cert.ListenAndServeTLS(server)
		} else {
			log.Print(`⚠ auth is running without secure connection ⚠`)
			serveError <- server.ListenAndServe()
		}
	}()

	httputils.ServerGracefulClose(server, serveError, nil)
}
