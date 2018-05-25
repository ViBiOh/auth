package main

import (
	"flag"

	"github.com/NYTimes/gziphandler"
	"github.com/ViBiOh/auth/pkg/provider/basic"
	"github.com/ViBiOh/auth/pkg/provider/github"
	"github.com/ViBiOh/auth/pkg/provider/twitter"
	"github.com/ViBiOh/auth/pkg/service"
	"github.com/ViBiOh/httputils/pkg"
	"github.com/ViBiOh/httputils/pkg/alcotest"
	"github.com/ViBiOh/httputils/pkg/cors"
	"github.com/ViBiOh/httputils/pkg/healthcheck"
	"github.com/ViBiOh/httputils/pkg/opentracing"
	"github.com/ViBiOh/httputils/pkg/owasp"
	"github.com/ViBiOh/httputils/pkg/server"
)

func main() {
	serverConfig := httputils.Flags(``)
	alcotestConfig := alcotest.Flags(``)
	opentracingConfig := opentracing.Flags(`tracing`)
	owaspConfig := owasp.Flags(``)
	corsConfig := cors.Flags(`cors`)

	serviceConfig := service.Flags(``)
	basicConfig := basic.Flags(`basic`)
	githubConfig := github.Flags(`github`)
	twitterConfig := twitter.Flags(`twitter`)

	flag.Parse()

	alcotest.DoAndExit(alcotestConfig)

	serverApp := httputils.NewApp(serverConfig)
	healthcheckApp := healthcheck.NewApp()
	opentracingApp := opentracing.NewApp(opentracingConfig)
	owaspApp := owasp.NewApp(owaspConfig)
	corsApp := cors.NewApp(corsConfig)

	serviceApp := service.NewApp(serviceConfig, basicConfig, githubConfig, twitterConfig)
	serviceHandler := server.ChainMiddlewares(gziphandler.GzipHandler(serviceApp.Handler()), opentracingApp, owaspApp, corsApp)

	serverApp.ListenAndServe(serviceHandler, nil, healthcheckApp)
}
