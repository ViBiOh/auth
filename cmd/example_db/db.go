package main

import (
	"flag"
	"net/http"
	"os"
	"strings"

	"github.com/ViBiOh/auth/v2/pkg/ident/basic"
	"github.com/ViBiOh/auth/v2/pkg/middleware"
	"github.com/ViBiOh/auth/v2/pkg/service"
	dbStore "github.com/ViBiOh/auth/v2/pkg/store/db"
	"github.com/ViBiOh/httputils/v3/pkg/crud"
	"github.com/ViBiOh/httputils/v3/pkg/db"
	"github.com/ViBiOh/httputils/v3/pkg/httputils"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
)

func main() {
	fs := flag.NewFlagSet("db", flag.ExitOnError)

	dbConfig := db.Flags(fs, "ident")
	crudConfig := crud.GetConfiguredFlags("auth", "")(fs, "ident")
	serverConfig := httputils.Flags(fs, "")

	logger.Fatal(fs.Parse(os.Args[1:]))

	appDB, err := db.New(dbConfig)
	logger.Fatal(err)

	authProvider := dbStore.New(appDB)
	identProvider := basic.New(authProvider)

	middlewareApp := middleware.New(authProvider, identProvider)

	crudHandler, err := crud.New(crudConfig, service.New(authProvider, authProvider))
	logger.Fatal(err)

	rawHandler := http.StripPrefix("/signup", crudHandler.Handler())
	protectedHandler := httputils.ChainMiddlewares(crudHandler.Handler(), middlewareApp.Middleware)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/signup") && r.Method == http.MethodPost {
			rawHandler.ServeHTTP(w, r)
		} else {
			protectedHandler.ServeHTTP(w, r)
		}
	})

	server := httputils.New(serverConfig)
	server.Health(appDB.Ping)
	server.ListenServeWait(handler)
}
