package main

import (
	"flag"
	"os"

	"github.com/ViBiOh/auth/v2/pkg/ident/basic"
	"github.com/ViBiOh/auth/v2/pkg/middleware"
	memoryStore "github.com/ViBiOh/auth/v2/pkg/store/memory"
	"github.com/ViBiOh/httputils/v4/pkg/health"
	"github.com/ViBiOh/httputils/v4/pkg/httputils"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/server"
	"github.com/ViBiOh/httputils/v4/pkg/tracer"
)

func main() {
	fs := flag.NewFlagSet("example", flag.ExitOnError)

	appServerConfig := server.Flags(fs, "")
	healthConfig := health.Flags(fs, "")

	loggerConfig := logger.Flags(fs, "logger")
	tracerConfig := tracer.Flags(fs, "tracer")

	basicConfig := memoryStore.Flags(fs, "")

	logger.Fatal(fs.Parse(os.Args[1:]))

	logger.Global(logger.New(loggerConfig))
	defer logger.Close()

	tracerApp, err := tracer.New(tracerConfig)
	logger.Fatal(err)
	defer tracerApp.Close()

	appServer := server.New(appServerConfig)
	healthApp := health.New(healthConfig)

	authProvider, err := memoryStore.New(basicConfig)
	logger.Fatal(err)

	identProvider := basic.New(authProvider, "Example Memory")
	middlewareApp := middleware.New(authProvider, tracerApp, identProvider)

	go appServer.Start("http", healthApp.End(), httputils.Handler(nil, healthApp, tracerApp.Middleware, middlewareApp.Middleware))

	healthApp.WaitForTermination(appServer.Done())
	server.GracefulWait(appServer.Done())
}
