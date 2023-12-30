package ident

import (
	"context"
	"errors"
	"net/http"

	"github.com/ViBiOh/auth/v2/pkg/model"
)

const MaxPasswordLength = 72

var (
	// ErrMalformedAuth occurs when authorization content is malformed
	ErrMalformedAuth = errors.New("malformed authorization content")

	// ErrUnavailableService occurs when service is unavailable
	ErrUnavailableService = errors.New("unavailable ident service")

	// ErrInvalidCredentials occurs when credentials failed
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrTooLongPassword occurs when password is too long
	ErrTooLongPassword = errors.New("password is too long, limit is 72 bytes")
)

// Provider provides methods for dealing with identification
type Provider interface {
	// IsMatching checks if header content match provider
	IsMatching(string) bool

	// GetUser returns User found in content header
	GetUser(context.Context, string) (model.User, error)

	// OnError handles HTTP request when login fails
	OnError(http.ResponseWriter, *http.Request, error)
}
