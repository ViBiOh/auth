package main

import (
	"flag"
	"os"

	authMemory "github.com/ViBiOh/auth/pkg/auth/memory"
	"github.com/ViBiOh/auth/pkg/handler"
	"github.com/ViBiOh/auth/pkg/ident/basic"
	basicMemory "github.com/ViBiOh/auth/pkg/ident/basic/memory"
	"github.com/ViBiOh/httputils/v3/pkg/httputils"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
)

func main() {
	fs := flag.NewFlagSet("api", flag.ExitOnError)

	basicMemoryConfig := basicMemory.Flags(fs, "ident")
	authMemoryConfig := authMemory.Flags(fs, "auth")

	serverConfig := httputils.Flags(fs, "")

	logger.Fatal(fs.Parse(os.Args[1:]))

	basicMemoryApp, err := basicMemory.New(basicMemoryConfig)
	logger.Fatal(err)

	authMemoryApp, err := authMemory.New(authMemoryConfig)
	logger.Fatal(err)

	basicMemoryProvider := basic.New(basicMemoryApp)
	handlerApp := handler.New(authMemoryApp, basicMemoryProvider)

	server := httputils.New(serverConfig)
	server.ListenServeWait(handlerApp.Handler(nil))
}
