package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/ViBiOh/alcotest/alcotest"
	"github.com/ViBiOh/auth/provider/basic"
	"github.com/ViBiOh/auth/provider/github"
	"github.com/ViBiOh/auth/service"
	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/httputils/cert"
	"github.com/ViBiOh/httputils/cors"
	"github.com/ViBiOh/httputils/owasp"
)

const (
	healthPrefix = `/health`
)

var (
	serviceHandler http.Handler
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			if _, err := w.Write(nil); err != nil {
				httputils.InternalServerError(w, err)
			}
			return
		}

		if r.Method == http.MethodGet && r.URL.Path == healthPrefix {
			healthHandler(w, r)
		} else {
			serviceHandler.ServeHTTP(w, r)
		}
	})
}

func main() {
	port := flag.String(`port`, `1080`, `Listen port`)
	tls := flag.Bool(`tls`, true, `Serve TLS content`)
	alcotestConfig := alcotest.Flags(``)
	certConfig := cert.Flags(`tls`)
	owaspConfig := owasp.Flags(``)
	corsConfig := cors.Flags(`cors`)

	serviceConfig := service.Flags(``)
	basicConfig := basic.Flags(`basic`)
	githubConfig := github.Flags(`github`)

	flag.Parse()

	alcotest.DoAndExit(alcotestConfig)

	log.Printf(`Starting server on port %s`, *port)

	serviceApp := service.NewApp(serviceConfig, basicConfig, githubConfig)
	serviceHandler = serviceApp.Handler()

	server := &http.Server{
		Addr:    `:` + *port,
		Handler: gziphandler.GzipHandler(owasp.Handler(owaspConfig, cors.Handler(corsConfig, handler()))),
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
