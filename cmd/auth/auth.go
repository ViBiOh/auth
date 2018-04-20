package main

import (
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/ViBiOh/auth/pkg/provider/basic"
	"github.com/ViBiOh/auth/pkg/provider/github"
	"github.com/ViBiOh/auth/pkg/provider/twitter"
	"github.com/ViBiOh/auth/pkg/service"
	"github.com/ViBiOh/httputils/pkg"
	"github.com/ViBiOh/httputils/pkg/cors"
	"github.com/ViBiOh/httputils/pkg/datadog"
	"github.com/ViBiOh/httputils/pkg/healthcheck"
	"github.com/ViBiOh/httputils/pkg/owasp"
)

const healthPrefix = `/health`

func main() {
	owaspConfig := owasp.Flags(``)
	corsConfig := cors.Flags(`cors`)
	serviceConfig := service.Flags(``)
	basicConfig := basic.Flags(`basic`)
	githubConfig := github.Flags(`github`)
	twitterConfig := twitter.Flags(`twitter`)
	datadogConfig := datadog.Flags(`datadog`)

	httputils.NewApp(httputils.Flags(``), func() http.Handler {
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

		return datadog.NewApp(datadogConfig).Handler(gziphandler.GzipHandler(owasp.Handler(owaspConfig, cors.Handler(corsConfig, handler))))
	}, nil).ListenAndServe()
}
