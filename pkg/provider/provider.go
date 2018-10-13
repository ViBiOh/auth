package provider

import (
	"context"
	"errors"
	"net/http"

	"github.com/ViBiOh/auth/pkg/model"
)

var (
	// ErrUnknownAuthType occurs when authentification type in unknown
	ErrUnknownAuthType = errors.New(`unknown authentication type`)

	// ErrInvalidState occurs when state is not consistent
	ErrInvalidState = errors.New(`invalid state provided for oauth`)

	// ErrInvalidCode occurs when code is no valid
	ErrInvalidCode = errors.New(`invalid code provided for oauth`)

	// ErrMalformedAuth occurs when auth header is malformed
	ErrMalformedAuth = errors.New(`malformed Authorization content`)
)

// Service provide methods for dealing with authentification
type Service interface {
	GetUser(context.Context, string) (*model.User, error)
}

// Auth is a provider of Authentification methods
type Auth interface {
	GetName() string
	GetUser(context.Context, string) (*model.User, error)
	OnUnauthorized(http.ResponseWriter, *http.Request, error)
	Redirect() (string, error)
	Login(*http.Request) (string, error)
}
