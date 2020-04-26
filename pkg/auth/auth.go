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
