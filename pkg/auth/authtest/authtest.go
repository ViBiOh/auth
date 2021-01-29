package authtest

import (
	"context"

	"github.com/ViBiOh/auth/v2/pkg/auth"
	"github.com/ViBiOh/auth/v2/pkg/model"
)

var _ auth.Provider = &App{}

// App mock app
type App struct {
	isAuthorized bool
}

// New creates raw mock
func New() *App {
	return &App{}
}

// SetIsAuthorized mocks
func (a *App) SetIsAuthorized(isAuthorized bool) *App {
	a.isAuthorized = isAuthorized

	return a
}

// IsAuthorized mocks
func (a *App) IsAuthorized(_ context.Context, _ model.User, _ string) bool {
	return a.isAuthorized
}
