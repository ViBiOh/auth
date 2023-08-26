package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"os"

	"github.com/ViBiOh/auth/v2/pkg/ident/basic"
	"github.com/ViBiOh/auth/v2/pkg/middleware"
	memoryStore "github.com/ViBiOh/auth/v2/pkg/store/memory"
	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/health"
	"github.com/ViBiOh/httputils/v4/pkg/httputils"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/server"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
)

func main() {
	fs := flag.NewFlagSet("example", flag.ExitOnError)
	fs.Usage = flags.Usage(fs)

	appServerConfig := server.Flags(fs, "")
	healthConfig := health.Flags(fs, "")

	loggerConfig := logger.Flags(fs, "logger")
	telemetryConfig := telemetry.Flags(fs, "telemetry")

	basicConfig := memoryStore.Flags(fs, "")

	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	logger.Init(loggerConfig)

	ctx := context.Background()

	telemetryService, err := telemetry.New(ctx, telemetryConfig)
	if err != nil {
		slog.Error("create tracer", "err", err)
		os.Exit(1)
	}

	defer telemetryService.Close(ctx)

	appServer := server.New(appServerConfig)
	healthService := health.New(healthConfig)

	authProvider, err := memoryStore.New(basicConfig)
	if err != nil {
		slog.Error("create memory store", "err", err)
		os.Exit(1)
	}

	identProvider := basic.New(authProvider, "Example Memory")
	middlewareApp := middleware.New(authProvider, telemetryService.TracerProvider(), identProvider)

	endCtx := healthService.End(ctx)

	go appServer.Start(endCtx, "http", httputils.Handler(nil, healthService, telemetryService.Middleware("http"), middlewareApp.Middleware))

	healthService.WaitForTermination(appServer.Done())
	server.GracefulWait(appServer.Done())
}
