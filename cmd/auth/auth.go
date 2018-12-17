package main

import (
	"flag"

	"github.com/ViBiOh/auth/pkg/ident"
	"github.com/ViBiOh/auth/pkg/ident/basic"
	"github.com/ViBiOh/auth/pkg/ident/github"
	"github.com/ViBiOh/auth/pkg/ident/handler"
	httputils "github.com/ViBiOh/httputils/pkg"
	"github.com/ViBiOh/httputils/pkg/alcotest"
	"github.com/ViBiOh/httputils/pkg/cors"
	"github.com/ViBiOh/httputils/pkg/gzip"
	"github.com/ViBiOh/httputils/pkg/healthcheck"
	"github.com/ViBiOh/httputils/pkg/logger"
	"github.com/ViBiOh/httputils/pkg/opentracing"
	"github.com/ViBiOh/httputils/pkg/owasp"
	"github.com/ViBiOh/httputils/pkg/prometheus"
	"github.com/ViBiOh/httputils/pkg/server"
)

func main() {
	fs := flag.NewFlagSet("auth", flag.ExitOnError)

	serverConfig := httputils.Flags(fs, ``)
	alcotestConfig := alcotest.Flags(fs, ``)
	prometheusConfig := prometheus.Flags(fs, `prometheus`)
	opentracingConfig := opentracing.Flags(fs, `tracing`)
	owaspConfig := owasp.Flags(fs, ``)
	corsConfig := cors.Flags(fs, `cors`)

	handlerConfig := handler.Flags(fs, ``)
	basicConfig := basic.Flags(fs, `basic`)
	githubConfig := github.Flags(fs, `github`)

	flag.Parse()

	alcotest.DoAndExit(alcotestConfig)

	serverApp := httputils.New(serverConfig)
	healthcheckApp := healthcheck.New()
	prometheusApp := prometheus.New(prometheusConfig)
	opentracingApp := opentracing.New(opentracingConfig)
	gzipApp := gzip.New()
	owaspApp := owasp.New(owaspConfig)
	corsApp := cors.New(corsConfig)

	basicApp, err := basic.New(basicConfig, nil)
	if err != nil {
		logger.Warn(`%+v`, err)
	}

	githubApp, err := github.New(githubConfig)
	if err != nil {
		logger.Warn(`%+v`, err)
	}

	identApp := handler.New(handlerConfig, []ident.Auth{basicApp, githubApp})
	identHandler := server.ChainMiddlewares(identApp.Handler(), prometheusApp, opentracingApp, gzipApp, owaspApp, corsApp)

	serverApp.ListenAndServe(identHandler, nil, healthcheckApp)
}
