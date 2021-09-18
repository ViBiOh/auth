package db

import (
	"context"

	"github.com/ViBiOh/auth/v2/pkg/auth"
	"github.com/ViBiOh/auth/v2/pkg/ident/basic"
	"github.com/jackc/pgx/v4"
)

// Database interface needed
type Database interface {
	Get(context.Context, func(pgx.Row) error, string, ...interface{}) error
	Create(context.Context, string, ...interface{}) (uint64, error)
	Exec(context.Context, string, ...interface{}) error
	DoAtomic(context.Context, func(context.Context) error) error
}

// App of package
type App struct {
	db Database
}

var (
	_ auth.Provider  = App{}
	_ auth.Storage   = App{}
	_ basic.Provider = App{}
)

// New creates new App from dependencies
func New(db Database) App {
	return App{
		db: db,
	}
}
