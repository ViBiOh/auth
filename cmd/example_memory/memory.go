package main

import (
	"context"
	"flag"
	"os"

	"github.com/ViBiOh/auth/v3/pkg/middleware"
	"github.com/ViBiOh/auth/v3/pkg/provider/basic"
	memoryStore "github.com/ViBiOh/auth/v3/pkg/store/memory"
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
	memoryConfig := memoryStore.Flags(fs, "")

	_ = fs.Parse(os.Args[1:])

	ctx := context.Background()

	logger.Init(ctx, loggerConfig)

	healthService := health.New(ctx, healthConfig)

	authProvider, err := memoryStore.New(memoryConfig)
	logger.FatalfOnErr(ctx, err, "create memory store")

	identProvider := basic.New(authProvider)
	middlewareApp := middleware.New(identProvider)

	appServer := server.New(serverConfig)
	go appServer.Start(healthService.EndCtx(), httputils.Handler(nil, healthService, middlewareApp.Middleware))

	healthService.WaitForTermination(appServer.Done())
	health.WaitAll(appServer.Done())
}
