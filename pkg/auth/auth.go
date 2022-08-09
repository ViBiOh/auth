package auth

import (
	"context"
	"errors"

	"github.com/ViBiOh/auth/v2/pkg/model"
)

//go:generate mockgen -source auth.go -destination ../mocks/auth.go -package mocks -mock_names Provider=Provider,Storage=Storage

// ErrForbidden occurs when user is identified but not authorized
var ErrForbidden = errors.New("forbidden access")

// Provider provides methods for dealing with identification
type Provider interface {
	// IsAuthorized checks if given user is authorized
	IsAuthorized(context.Context, model.User, string) bool
}

// Storage defines interaction with storage from User
type Storage interface {
	DoAtomic(context.Context, func(context.Context) error) error
	Get(context.Context, uint64) (model.User, error)
	Create(context.Context, model.User) (uint64, error)
	Update(context.Context, model.User) error
	Delete(context.Context, model.User) error
}

// Service defines interaction with storage and provider from User
type Service interface {
	Get(context.Context, uint64) (model.User, error)
	Create(context.Context, model.User) (model.User, error)
	Update(context.Context, model.User) (model.User, error)
	Delete(context.Context, model.User) error
	Check(context.Context, model.User, model.User) error
}
