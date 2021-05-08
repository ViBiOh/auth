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
)

func main() {
	fs := flag.NewFlagSet("example", flag.ExitOnError)

	appServerConfig := server.Flags(fs, "")
	healthConfig := health.Flags(fs, "")

	basicConfig := memoryStore.Flags(fs, "")

	logger.Fatal(fs.Parse(os.Args[1:]))

	appServer := server.New(appServerConfig)
	healthApp := health.New(healthConfig)

	authProvider, err := memoryStore.New(basicConfig)
	logger.Fatal(err)

	identProvider := basic.New(authProvider, "Example Memory")
	middlewareApp := middleware.New(authProvider, identProvider)

	go appServer.Start("http", healthApp.End(), httputils.Handler(nil, healthApp, middlewareApp.Middleware))

	healthApp.WaitForTermination(appServer.Done())
	server.GracefulWait(appServer.Done())
}
