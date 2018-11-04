package ident

import (
	"context"
	"errors"
	"net/http"

	"github.com/ViBiOh/auth/pkg/model"
)

var (
	// ErrUnknownIdentType occurs when identification type in unknown
	ErrUnknownIdentType = errors.New(`unknown identification type`)

	// ErrEmptyAuth occurs when authorization content is not found
	ErrEmptyAuth = errors.New(`empty authorization content`)

	// ErrMalformedAuth occurs when authorization content is malformed
	ErrMalformedAuth = errors.New(`malformed authorization content`)

	// ErrInvalidState occurs when state is not consistent
	ErrInvalidState = errors.New(`invalid state provided for oauth`)

	// ErrInvalidCode occurs when code is no valid
	ErrInvalidCode = errors.New(`invalid code provided for oauth`)
)

// Service provide methods for dealing with identification
type Service interface {
	GetUser(context.Context, string) (*model.User, error)
	OnError(http.ResponseWriter, *http.Request, error)
}

// Auth is a provider of identification methods
type Auth interface {
	GetName() string
	GetUser(context.Context, string) (*model.User, error)
	Redirect(http.ResponseWriter, *http.Request)
	Login(*http.Request) (string, error)
	OnLoginError(http.ResponseWriter, *http.Request, error)
}
