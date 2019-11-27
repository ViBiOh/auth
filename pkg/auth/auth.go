package auth

import (
	"errors"

	"github.com/ViBiOh/auth/pkg/model"
)

var (
	// ErrForbidden occurs when user is identified but not authorized
	ErrForbidden = errors.New("forbidden access")
)

// Provider provides methods for dealing with identification
type Provider interface {
	// IsAuthorized checks if given user is authorized
	IsAuthorized(model.User, string) bool
}
