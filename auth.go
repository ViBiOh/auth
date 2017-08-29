package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"

	"github.com/ViBiOh/alcotest/alcotest"
	"github.com/ViBiOh/auth/basic"
	"github.com/ViBiOh/auth/github"
	"github.com/ViBiOh/auth/rate"
	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/httputils/cert"
	"github.com/ViBiOh/httputils/cors"
	"github.com/ViBiOh/httputils/owasp"
	"github.com/ViBiOh/httputils/prometheus"
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

func authHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Write(nil)
		return
	}

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if !rate.CheckRate(r) {
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}

	if r.URL.Path == `/user` {
		userHandler(w, r)
	} else if r.URL.Path == `/token/github` {
		githubTokenHandler(w, r)
	} else if r.URL.Path == `/health` {
		healthHandler(w, r)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func main() {
	url := flag.String(`c`, ``, `URL to healthcheck (check and exit)`)
	port := flag.String(`port`, `1080`, `Listen port`)
	flag.Parse()

	if *url != `` {
		alcotest.Do(url)
		return
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	log.Printf(`Starting server on port %s`, *port)

	Init()

	server := &http.Server{
		Addr:    `:` + *port,
		Handler: prometheus.NewPrometheusHandler(`http`, owasp.Handler{Handler: cors.Handler{Handler: http.HandlerFunc(authHandler)}}),
	}

	go log.Panic(cert.ListenAndServeTLS(server))
	httputils.ServerGracefulClose(server, nil)
}
