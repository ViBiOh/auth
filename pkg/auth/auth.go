package auth

import (
	"context"
	"errors"

	"github.com/ViBiOh/auth/v2/pkg/model"
)

var (
	// ErrForbidden occurs when user is identified but not authorized
	ErrForbidden = errors.New("forbidden access")
)

// Provider provides methods for dealing with identification
type Provider interface {
	// IsAuthorized checks if given user is authorized
	IsAuthorized(context.Context, model.User, string) bool
}

// Storage defines interaction with storage from User
type Storage interface {
	DoAtomic(ctx context.Context, action func(context.Context) error) error
	Get(ctx context.Context, id uint64) (model.User, error)
	Create(ctx context.Context, o model.User) (uint64, error)
	Update(ctx context.Context, o model.User) error
	Delete(ctx context.Context, o model.User) error
}
