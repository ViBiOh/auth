package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/ViBiOh/auth/v2/pkg/middleware"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/auth/v2/pkg/provider/github"
	dbStore "github.com/ViBiOh/auth/v2/pkg/store/db"
	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/db"
	"github.com/ViBiOh/httputils/v4/pkg/health"
	"github.com/ViBiOh/httputils/v4/pkg/httperror"
	"github.com/ViBiOh/httputils/v4/pkg/httputils"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/redis"
	"github.com/ViBiOh/httputils/v4/pkg/server"
)

func main() {
	fs := flag.NewFlagSet("oauth", flag.ExitOnError)
	fs.Usage = flags.Usage(fs)

	loggerConfig := logger.Flags(fs, "logger")
	healthConfig := health.Flags(fs, "")

	serverConfig := server.Flags(fs, "")
	redisConfig := redis.Flags(fs, "redis")
	githubConfig := github.Flags(fs, "github")
	dbConfig := db.Flags(fs, "db")

	_ = fs.Parse(os.Args[1:])

	ctx := context.Background()

	logger.Init(ctx, loggerConfig)

	healthService := health.New(ctx, healthConfig)

	redisClient, err := redis.New(ctx, redisConfig, nil, nil)
	logger.FatalfOnErr(ctx, err, "redis")

	dbProvider, err := db.New(ctx, dbConfig, nil)
	logger.FatalfOnErr(ctx, err, "create db provider")

	githubService := github.New(githubConfig, redisClient, dbStore.New(dbProvider))
	authMiddleware := middleware.New(githubService, "", nil)

	authMux := http.NewServeMux()
	authMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		payload, err := json.Marshal(model.ReadUser(r.Context()))
		if err != nil {
			httperror.InternalServerError(r.Context(), w, err)
			return
		}

		fmt.Fprintf(w, "%s", payload)
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/auth/github/callback", githubService.Callback)
	mux.Handle("/", authMiddleware.Middleware(authMux))

	appServer := server.New(serverConfig)
	go appServer.Start(healthService.EndCtx(), httputils.Handler(mux, healthService))

	healthService.WaitForTermination(appServer.Done())
	health.WaitAll(appServer.Done())
}
