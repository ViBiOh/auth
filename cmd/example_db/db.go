package main

import (
	"flag"
	"net/http"
	"os"
	"strings"

	"github.com/ViBiOh/auth/v2/pkg/handler"
	"github.com/ViBiOh/auth/v2/pkg/ident/basic"
	basicDb "github.com/ViBiOh/auth/v2/pkg/provider/db"
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

	basicApp := basicDb.New(appDB)
	basicProvider := basic.New(basicApp)
	handlerApp := handler.New(basicApp, basicProvider)

	crudHandler, err := crud.New(crudConfig, service.New(appDB, basicApp))
	logger.Fatal(err)

	rawHandler := http.StripPrefix("/signup", crudHandler.Handler())
	protectedHandler := httputils.ChainMiddlewares(crudHandler.Handler(), handlerApp.Middleware)

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
