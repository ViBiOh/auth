package db

import (
	"context"

	"github.com/ViBiOh/auth/v2/pkg/auth"
	"github.com/ViBiOh/auth/v2/pkg/ident/basic"
	"github.com/jackc/pgx/v5"
)

//go:generate mockgen -source db.go -destination ../../mocks/db.go -package mocks -mock_names Database=Database

type Database interface {
	Get(context.Context, func(pgx.Row) error, string, ...any) error
	Create(context.Context, string, ...any) (uint64, error)
	One(context.Context, string, ...any) error
	DoAtomic(context.Context, func(context.Context) error) error
}

type Service struct {
	db Database
}

var (
	_ auth.Provider       = Service{}
	_ auth.Storage        = Service{}
	_ basic.LoginProvider = Service{}
)

func New(db Database) Service {
	return Service{
		db: db,
	}
}
