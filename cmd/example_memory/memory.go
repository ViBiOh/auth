package main

import (
	"flag"
	"os"

	auth "github.com/ViBiOh/auth/v2/pkg/auth/memory"
	"github.com/ViBiOh/auth/v2/pkg/handler"
	"github.com/ViBiOh/auth/v2/pkg/ident/basic"
	basicProvider "github.com/ViBiOh/auth/v2/pkg/ident/basic/memory"
	"github.com/ViBiOh/httputils/v3/pkg/httputils"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
)

func main() {
	fs := flag.NewFlagSet("memory", flag.ExitOnError)

	basicProviderConfig := basicProvider.Flags(fs, "ident")
	authConfig := auth.Flags(fs, "auth")

	serverConfig := httputils.Flags(fs, "")

	logger.Fatal(fs.Parse(os.Args[1:]))

	basicProviderApp, err := basicProvider.New(basicProviderConfig)
	logger.Fatal(err)

	authApp, err := auth.New(authConfig)
	logger.Fatal(err)

	basicProviderProvider := basic.New(basicProviderApp)
	handlerApp := handler.New(authApp, basicProviderProvider)

	server := httputils.New(serverConfig)
	server.ListenServeWait(handlerApp.Handler(nil))
}
