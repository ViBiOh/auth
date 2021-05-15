package db

import (
	"context"
	"database/sql"

	"github.com/ViBiOh/auth/v2/pkg/auth"
	"github.com/ViBiOh/auth/v2/pkg/ident/basic"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/auth/v2/pkg/store"
)

// App of package
type App interface {
	DoAtomic(ctx context.Context, action func(context.Context) error) error

	Get(ctx context.Context, id uint64) (model.User, error)
	Create(ctx context.Context, o model.User) (uint64, error)
	Update(ctx context.Context, o model.User) error
	Delete(ctx context.Context, o model.User) error

	Login(ctx context.Context, login, password string) (model.User, error)
	IsAuthorized(ctx context.Context, user model.User, profile string) bool
}

var (
	_ auth.Provider     = app{}
	_ basic.Provider    = app{}
	_ store.UserStorage = app{}
)

type app struct {
	db *sql.DB
}

// New creates new App from dependencies
func New(db *sql.DB) App {
	return app{
		db: db,
	}
}
