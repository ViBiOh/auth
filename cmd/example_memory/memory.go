package main

import (
	"flag"
	"os"

	"github.com/ViBiOh/auth/v2/pkg/handler"
	"github.com/ViBiOh/auth/v2/pkg/ident/basic"
	basicMemory "github.com/ViBiOh/auth/v2/pkg/provider/memory"
	"github.com/ViBiOh/httputils/v3/pkg/httputils"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
)

func main() {
	fs := flag.NewFlagSet("memory", flag.ExitOnError)

	basicConfig := basicMemory.Flags(fs, "")
	serverConfig := httputils.Flags(fs, "")

	logger.Fatal(fs.Parse(os.Args[1:]))

	basicApp, err := basicMemory.New(basicConfig)
	logger.Fatal(err)

	basicProviderProvider := basic.New(basicApp)
	handlerApp := handler.New(basicApp, basicProviderProvider)

	server := httputils.New(serverConfig)
	server.ListenServeWait(handlerApp.Middleware(nil))
}
