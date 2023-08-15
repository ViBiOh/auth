package main

import (
	"context"
	"flag"
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
	"github.com/ViBiOh/httputils/v4/pkg/tracer"
)

func main() {
	fs := flag.NewFlagSet("example", flag.ExitOnError)
	fs.Usage = flags.Usage(fs)

	appServerConfig := server.Flags(fs, "")
	healthConfig := health.Flags(fs, "")

	loggerConfig := logger.Flags(fs, "logger")
	tracerConfig := tracer.Flags(fs, "tracer")

	basicConfig := memoryStore.Flags(fs, "")

	if err := fs.Parse(os.Args[1:]); err != nil {
		slog.Error("parse flags", "err", err)
		os.Exit(1)
	}

	logger.New(loggerConfig)

	ctx := context.Background()

	tracerApp, err := tracer.New(ctx, tracerConfig)
	if err != nil {
		slog.Error("create tracer", "err", err)
		os.Exit(1)
	}

	defer tracerApp.Close(ctx)

	appServer := server.New(appServerConfig)
	healthApp := health.New(healthConfig)

	authProvider, err := memoryStore.New(basicConfig)
	if err != nil {
		slog.Error("create memory store", "err", err)
		os.Exit(1)
	}

	identProvider := basic.New(authProvider, "Example Memory")
	middlewareApp := middleware.New(authProvider, tracerApp.GetTracer("auth"), identProvider)

	endCtx := healthApp.End(ctx)

	go appServer.Start(endCtx, "http", httputils.Handler(nil, healthApp, tracerApp.Middleware, middlewareApp.Middleware))

	healthApp.WaitForTermination(appServer.Done())
	server.GracefulWait(appServer.Done())
}
