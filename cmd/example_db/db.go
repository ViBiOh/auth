package main

import (
	"flag"
	"os"

	auth "github.com/ViBiOh/auth/pkg/auth/db"
	"github.com/ViBiOh/auth/pkg/handler"
	"github.com/ViBiOh/auth/pkg/ident/basic"
	basicProvider "github.com/ViBiOh/auth/pkg/ident/basic/db"
	"github.com/ViBiOh/httputils/v3/pkg/db"
	"github.com/ViBiOh/httputils/v3/pkg/httputils"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
)

func main() {
	fs := flag.NewFlagSet("db", flag.ExitOnError)

	dbConfig := db.Flags(fs, "ident")

	serverConfig := httputils.Flags(fs, "")

	logger.Fatal(fs.Parse(os.Args[1:]))

	appDB, err := db.New(dbConfig)
	logger.Fatal(err)

	basicApp := basicProvider.New(appDB)
	authApp := auth.New(appDB)

	basicProvider := basic.New(basicApp)
	handlerApp := handler.New(authApp, basicProvider)

	server := httputils.New(serverConfig)
	server.ListenServeWait(handlerApp.Handler(nil))
}
