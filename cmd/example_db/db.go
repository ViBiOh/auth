package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"os"

	"github.com/ViBiOh/auth/v2/pkg/ident/basic"
	"github.com/ViBiOh/auth/v2/pkg/middleware"
	dbStore "github.com/ViBiOh/auth/v2/pkg/store/db"
	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/db"
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

	dbConfig := db.Flags(fs, "db")

	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	logger.Init(loggerConfig)

	ctx := context.Background()

	telemetryApp, err := telemetry.New(ctx, telemetryConfig)
	if err != nil {
		slog.Error("create tracer", "err", err)
		os.Exit(1)
	}

	defer telemetryApp.Close(ctx)

	appServer := server.New(appServerConfig)

	appDB, err := db.New(ctx, dbConfig, nil)
	if err != nil {
		slog.Error("create db", "err", err)
		os.Exit(1)
	}

	defer appDB.Close()

	healthApp := health.New(healthConfig, appDB.Ping)

	authProvider := dbStore.New(appDB)
	identProvider := basic.New(authProvider, "Example with a DB")
	middlewareApp := middleware.New(authProvider, telemetryApp.GetTracer("auth"), identProvider)

	endCtx := healthApp.End(ctx)

	go appServer.Start(endCtx, "http", httputils.Handler(nil, healthApp, telemetryApp.Middleware("http"), middlewareApp.Middleware))

	healthApp.WaitForTermination(appServer.Done())
	server.GracefulWait(appServer.Done())
}
