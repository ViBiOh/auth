package service

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"github.com/ViBiOh/auth/pkg/ident"
	"github.com/ViBiOh/auth/pkg/ident/basic"
	"github.com/ViBiOh/auth/pkg/model"
	"github.com/ViBiOh/httputils/pkg/logger"
)

// App stores informations and secret of API
type App struct {
	provider ident.Auth
}

// NewBasic creates new App from Flags' config only for basic auth wrapper
func NewBasic(basicConfig basic.Config, db *sql.DB) *App {
	provider, err := basic.New(basicConfig, db)
	if err != nil {
		logger.Fatal("%#v", err)
	}

	return &App{
		provider: provider,
	}
}

// GetUser get user from given auth content
func (a App) GetUser(ctx context.Context, authContent string) (*model.User, error) {
	if authContent == "" {
		return nil, ident.ErrEmptyAuth
	}

	parts := strings.SplitN(authContent, " ", 2)
	if len(parts) != 2 {
		return nil, ident.ErrMalformedAuth
	}

	if parts[0] != a.provider.GetName() {
		return nil, ident.ErrUnknownIdentType
	}

	return a.provider.GetUser(ctx, parts[1])
}

// OnError handle error for service app
func (a App) OnError(w http.ResponseWriter, r *http.Request, err error) {
	a.provider.OnLoginError(w, r, err)
}
