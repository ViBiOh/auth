package main

import (
	"flag"
	"os"

	"github.com/ViBiOh/auth/pkg/handler"
	"github.com/ViBiOh/auth/pkg/ident"
	"github.com/ViBiOh/auth/pkg/ident/basic"
	"github.com/ViBiOh/auth/pkg/ident/basic/memory"
	"github.com/ViBiOh/httputils/v3/pkg/httputils"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
)

func main() {
	fs := flag.NewFlagSet("api", flag.ExitOnError)

	basicMemoryConfig := memory.Flags(fs, "memory")

	serverConfig := httputils.Flags(fs, "")

	logger.Fatal(fs.Parse(os.Args[1:]))

	basicMemoryApp, err := memory.New(basicMemoryConfig)
	logger.Fatal(err)

	basicMemoryProvider := basic.New(basicMemoryApp)
	handlerApp := handler.New([]ident.Provider{basicMemoryProvider})

	server := httputils.New(serverConfig)
	server.ListenServeWait(handlerApp.Handler(nil))
}
