package main

import (
	"flag"
	"net/http"
	"os"
	"strings"

	auth "github.com/ViBiOh/auth/v2/pkg/auth/db"
	"github.com/ViBiOh/auth/v2/pkg/handler"
	"github.com/ViBiOh/auth/v2/pkg/ident/basic"
	basicProvider "github.com/ViBiOh/auth/v2/pkg/ident/basic/db"
	"github.com/ViBiOh/auth/v2/pkg/service"
	"github.com/ViBiOh/httputils/v3/pkg/crud"
	"github.com/ViBiOh/httputils/v3/pkg/db"
	"github.com/ViBiOh/httputils/v3/pkg/httputils"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
)

func main() {
	fs := flag.NewFlagSet("db", flag.ExitOnError)

	dbConfig := db.Flags(fs, "ident")
	crudConfig := crud.Flags(fs, "ident")

	serverConfig := httputils.Flags(fs, "")

	logger.Fatal(fs.Parse(os.Args[1:]))

	appDB, err := db.New(dbConfig)
	logger.Fatal(err)

	basicApp := basicProvider.New(appDB)
	authApp := auth.New(appDB)

	basicProvider := basic.New(basicApp)
	handlerApp := handler.New(authApp, basicProvider)

	crudHandler, err := crud.New(crudConfig, service.New(appDB, authApp))
	logger.Fatal(err)

	rawHandler := http.StripPrefix("/signup", crudHandler.Handler())
	protectedHandler := httputils.ChainMiddlewares(crudHandler.Handler(), handlerApp)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/signup") && r.Method == http.MethodPost {
			rawHandler.ServeHTTP(w, r)
		} else {
			protectedHandler.ServeHTTP(w, r)
		}
	})

	server := httputils.New(serverConfig)
	server.ListenServeWait(handler)
}
