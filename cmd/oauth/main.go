package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/ViBiOh/auth/v2/pkg/ident/github"
	"github.com/ViBiOh/auth/v2/pkg/middleware"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/auth/v2/pkg/store/memory"
	"github.com/ViBiOh/flags"
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
	memoryConfig := memory.Flags(fs, "memory")

	_ = fs.Parse(os.Args[1:])

	ctx := context.Background()

	logger.Init(ctx, loggerConfig)

	healthService := health.New(ctx, healthConfig)

	redisClient, err := redis.New(ctx, redisConfig, nil, nil)
	logger.FatalfOnErr(ctx, err, "redis")

	authProvider, err := memory.New(memoryConfig)
	logger.FatalfOnErr(ctx, err, "memory")

	githubService := github.New(githubConfig, redisClient)
	authMiddleware := middleware.New(authProvider, nil, githubService)

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
