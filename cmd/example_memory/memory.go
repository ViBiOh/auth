package main

import (
	"flag"
	"os"

	"github.com/ViBiOh/auth/v2/pkg/ident/basic"
	"github.com/ViBiOh/auth/v2/pkg/middleware"
	memoryStore "github.com/ViBiOh/auth/v2/pkg/store/memory"
	"github.com/ViBiOh/httputils/v3/pkg/httputils"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
)

func main() {
	fs := flag.NewFlagSet("memory", flag.ExitOnError)

	basicConfig := memoryStore.Flags(fs, "")
	serverConfig := httputils.Flags(fs, "")

	logger.Fatal(fs.Parse(os.Args[1:]))

	authProvider, err := memoryStore.New(basicConfig)
	logger.Fatal(err)

	identProvider := basic.New(authProvider)
	middlewareApp := middleware.New(authProvider, identProvider)

	server := httputils.New(serverConfig)
	server.ListenServeWait(middlewareApp.Middleware(nil))
}
