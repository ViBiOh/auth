package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/ViBiOh/auth/v2/pkg/ident/basic"
	"github.com/ViBiOh/auth/v2/pkg/middleware"
	"github.com/ViBiOh/auth/v2/pkg/service"
	dbStore "github.com/ViBiOh/auth/v2/pkg/store/db"
	"github.com/ViBiOh/httputils/v3/pkg/crud"
	"github.com/ViBiOh/httputils/v3/pkg/db"
	"github.com/ViBiOh/httputils/v3/pkg/httputils"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/httputils/v3/pkg/query"
)

func main() {
	fs := flag.NewFlagSet("example", flag.ExitOnError)

	dbConfig := db.Flags(fs, "db")
	crudConfig := crud.GetConfiguredFlags("/users", "User")(fs, "crud")
	serverConfig := httputils.Flags(fs, "")

	logger.Fatal(fs.Parse(os.Args[1:]))

	appDB, err := db.New(dbConfig)
	logger.Fatal(err)

	authProvider := dbStore.New(appDB)
	identProvider := basic.New(authProvider)
	middlewareApp := middleware.New(authProvider, identProvider)

	crudApp, err := crud.New(crudConfig, service.New(authProvider, authProvider))
	logger.Fatal(err)

	crudHandler := crudApp.Handler()

	protectedHandler := httputils.ChainMiddlewares(crudHandler, middlewareApp.Middleware)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && query.IsRoot(r) {
			crudHandler.ServeHTTP(w, r)
			return
		}

		protectedHandler.ServeHTTP(w, r)
	})

	server := httputils.New(serverConfig)
	server.Health(appDB.Ping)
	server.ListenServeWait(handler)
}
