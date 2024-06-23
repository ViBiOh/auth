package main

import (
	"context"
	"flag"
	"os"

	"github.com/ViBiOh/auth/v2/pkg/ident/basic"
	"github.com/ViBiOh/auth/v2/pkg/middleware"
	memoryStore "github.com/ViBiOh/auth/v2/pkg/store/memory"
	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/health"
	"github.com/ViBiOh/httputils/v4/pkg/httputils"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/server"
)

func main() {
	fs := flag.NewFlagSet("example", flag.ExitOnError)
	fs.Usage = flags.Usage(fs)

	loggerConfig := logger.Flags(fs, "logger")
	healthConfig := health.Flags(fs, "")

	serverConfig := server.Flags(fs, "")
	basicConfig := memoryStore.Flags(fs, "")

	_ = fs.Parse(os.Args[1:])

	ctx := context.Background()

	logger.Init(ctx, loggerConfig)

	healthService := health.New(ctx, healthConfig)

	authProvider, err := memoryStore.New(basicConfig)
	logger.FatalfOnErr(ctx, err, "create memory store")

	identProvider := basic.New(authProvider, "Example Memory")
	middlewareApp := middleware.New(authProvider, nil, identProvider)

	appServer := server.New(serverConfig)
	go appServer.Start(healthService.EndCtx(), httputils.Handler(nil, healthService, middlewareApp.Middleware))

	healthService.WaitForTermination(appServer.Done())
	server.GracefulWait(appServer.Done())
}
