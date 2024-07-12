package main

import (
	"context"
	"flag"
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
)

func main() {
	fs := flag.NewFlagSet("example", flag.ExitOnError)
	fs.Usage = flags.Usage(fs)

	loggerConfig := logger.Flags(fs, "logger")
	healthConfig := health.Flags(fs, "")

	serverConfig := server.Flags(fs, "")
	dbConfig := db.Flags(fs, "db")

	_ = fs.Parse(os.Args[1:])

	ctx := context.Background()

	logger.Init(ctx, loggerConfig)

	appDB, err := db.New(ctx, dbConfig, nil)
	logger.FatalfOnErr(ctx, err, "create db")

	defer appDB.Close()

	healthService := health.New(ctx, healthConfig, appDB.Ping)

	authProvider := dbStore.New(appDB)
	identProvider := basic.New(authProvider, "Example with a DB")
	middlewareApp := middleware.New(authProvider, nil, identProvider)

	appServer := server.New(serverConfig)
	go appServer.Start(healthService.EndCtx(), httputils.Handler(nil, healthService, middlewareApp.Middleware))

	healthService.WaitForTermination(appServer.Done())
	health.WaitAll(appServer.Done())
}
