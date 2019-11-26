package ident

import (
	"context"
	"errors"
	"net/http"

	"github.com/ViBiOh/auth/pkg/model"
)

var (
	// ErrEmptyAuth occurs when authorization content is not found
	ErrEmptyAuth = errors.New("empty authorization content")

	// ErrMalformedAuth occurs when authorization content is malformed
	ErrMalformedAuth = errors.New("malformed authorization content")

	// ErrUnavailableService occurs when service is unavailable
	ErrUnavailableService = errors.New("unavailable ident service")

	// ErrInvalidCredentials occurs when credentials failed
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Provider provide methods for dealing with identification
type Provider interface {
	GetUser(context.Context, string) (model.User, error)
	OnError(http.ResponseWriter, *http.Request, error)
}
