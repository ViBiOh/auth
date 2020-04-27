package store

import (
	"context"

	"github.com/ViBiOh/auth/v2/pkg/model"
)

// UserStorage defines interaction with storage from User Service
type UserStorage interface {
	List(ctx context.Context, page, pageSize uint, sortKey string, sortAsc bool) ([]model.User, uint, error)
	Get(ctx context.Context, id uint64) (model.User, error)
	Create(ctx context.Context, o model.User) (uint64, error)
	Update(ctx context.Context, o model.User) error
	Delete(ctx context.Context, o model.User) error
}
