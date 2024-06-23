package ident

import (
	"context"
	"errors"
	"net/http"

	"github.com/ViBiOh/auth/v2/pkg/model"
)

var (
	ErrMalformedAuth      = errors.New("malformed authorization content")
	ErrUnavailableService = errors.New("unavailable ident service")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTooLongPassword    = errors.New("password is too long, limit is 72 bytes")
)

// Provider provides methods for dealing with identification
type Provider interface {
	IsMatching(string) bool
	GetUser(context.Context, string) (model.User, error)
	OnError(http.ResponseWriter, *http.Request, error)
}
