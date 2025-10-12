package db

import (
	"context"

	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/auth/v2/pkg/provider/basic"
	"github.com/jackc/pgx/v5"
)

//go:generate mockgen -source $GOFILE -destination ../../mocks/$GOFILE -package mocks -mock_names Database=Database

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
	_ model.Storage  = Service{}
	_ basic.Provider = Service{}
)

func New(db Database) Service {
	return Service{
		db: db,
	}
}
