package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/ViBiOh/auth/v3/pkg/middleware"
	"github.com/ViBiOh/auth/v3/pkg/model"
	"github.com/ViBiOh/auth/v3/pkg/provider/github"
	dbStore "github.com/ViBiOh/auth/v3/pkg/store/db"
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

	database, err := db.New(ctx, dbConfig, nil)
	logger.FatalfOnErr(ctx, err, "create database")

	dbService := dbStore.New(database)

	var registration string
	err = dbService.DoAtomic(ctx, func(ctx context.Context) error {
		user, err := dbService.Create(ctx)
		if err != nil {
			return fmt.Errorf("create user: %w", err)
		}

		user.Name = "ViBiOh"

		registration, err = dbService.CreateGitHubRegistration(ctx, user)
		if err != nil {
			return fmt.Errorf("create github registration: %w", err)
		}

		return nil
	})
	logger.FatalfOnErr(ctx, err, "do atomic ")

	fmt.Printf("Connect to http://127.0.0.1:%d/auth/github/register?registration=%s\n", serverConfig.Port, registration)

	githubService := github.New(githubConfig, redisClient, dbService)
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
	mux.HandleFunc("/auth/github/register", githubService.Register)
	mux.Handle("/", authMiddleware.Middleware(authMux))

	appServer := server.New(serverConfig)
	go appServer.Start(healthService.EndCtx(), httputils.Handler(mux, healthService))

	healthService.WaitForTermination(appServer.Done())
	health.WaitAll(appServer.Done())
}
