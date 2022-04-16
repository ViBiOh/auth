package main

import (
	"flag"
	"os"

	"github.com/ViBiOh/auth/v2/pkg/ident/basic"
	"github.com/ViBiOh/auth/v2/pkg/middleware"
	dbStore "github.com/ViBiOh/auth/v2/pkg/store/db"
	"github.com/ViBiOh/httputils/v4/pkg/db"
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

	dbConfig := db.Flags(fs, "db")

	logger.Fatal(fs.Parse(os.Args[1:]))

	logger.Global(logger.New(loggerConfig))
	defer logger.Close()

	tracerApp, err := tracer.New(tracerConfig)
	logger.Fatal(err)
	defer tracerApp.Close()

	appServer := server.New(appServerConfig)

	appDB, err := db.New(dbConfig, nil)
	logger.Fatal(err)
	defer appDB.Close()

	healthApp := health.New(healthConfig, appDB.Ping)

	authProvider := dbStore.New(appDB)
	identProvider := basic.New(authProvider, "Example with a DB")
	middlewareApp := middleware.New(authProvider, tracerApp, identProvider)

	go appServer.Start("http", healthApp.End(), httputils.Handler(nil, healthApp, tracerApp.Middleware, middlewareApp.Middleware))

	healthApp.WaitForTermination(appServer.Done())
	server.GracefulWait(appServer.Done())
}
