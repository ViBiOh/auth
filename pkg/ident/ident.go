package ident

import (
	"errors"
	"net/http"

	"github.com/ViBiOh/auth/pkg/model"
)

var (
	// ErrMalformedAuth occurs when authorization content is malformed
	ErrMalformedAuth = errors.New("malformed authorization content")

	// ErrUnavailableService occurs when service is unavailable
	ErrUnavailableService = errors.New("unavailable ident service")

	// ErrInvalidCredentials occurs when credentials failed
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Provider provides methods for dealing with identification
type Provider interface {
	// IsMatching checks if header content match provider
	IsMatching(string) bool

	// GetUser returns User found in content header
	GetUser(string) (model.User, error)

	// OnError handles HTTP request when login fails
	OnError(http.ResponseWriter, *http.Request, error)
}
