package main

import (
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/ViBiOh/auth/provider/basic"
	"github.com/ViBiOh/auth/provider/github"
	"github.com/ViBiOh/auth/provider/twitter"
	"github.com/ViBiOh/auth/service"
	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/httputils/cors"
	"github.com/ViBiOh/httputils/healthcheck"
	"github.com/ViBiOh/httputils/owasp"
)

const healthPrefix = `/health`

func main() {
	owaspConfig := owasp.Flags(``)
	corsConfig := cors.Flags(`cors`)
	serviceConfig := service.Flags(``)
	basicConfig := basic.Flags(`basic`)
	githubConfig := github.Flags(`github`)
	twitterConfig := twitter.Flags(`twitter`)

	httputils.StartMainServer(func() http.Handler {
		serviceApp := service.NewApp(serviceConfig, basicConfig, githubConfig, twitterConfig)
		serviceHandler := serviceApp.Handler()

		healthHandler := http.StripPrefix(healthPrefix, healthcheck.Handler())

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == healthPrefix {
				healthHandler.ServeHTTP(w, r)
			} else {
				serviceHandler.ServeHTTP(w, r)
			}
		})

		return gziphandler.GzipHandler(owasp.Handler(owaspConfig, cors.Handler(corsConfig, handler)))
	}, nil)
}
