package provider

import (
	"context"
	"errors"
	"net/http"

	"github.com/ViBiOh/auth/pkg/model"
)

var (
	// ErrInvalidState occurs when state is not consistent
	ErrInvalidState = errors.New(`invalid state provided for oauth`)

	// ErrInvalidCode occurs when code is no valid
	ErrInvalidCode = errors.New(`invalid code provided for oauth`)

	// ErrUnknownAuthType occurs when authentification type in unknown
	ErrUnknownAuthType = errors.New(`unknown authentication type`)

	// ErrEmptyAuthorization occurs when authorization content is not found
	ErrEmptyAuthorization = errors.New(`empty authorization content`)

	// ErrMalformedAuth occurs when auth header is malformed
	ErrMalformedAuth = errors.New(`malformed Authorization content`)

	// ErrForbiden occurs when user is authentified but not granted
	ErrForbiden = errors.New(`forbidden access`)
)

// Service provide methods for dealing with authentification
type Service interface {
	GetUser(context.Context, string) (*model.User, error)
	RedirectToFirstProvider(http.ResponseWriter, *http.Request) bool
}

// Auth is a provider of Authentification methods
type Auth interface {
	GetName() string
	GetUser(context.Context, string) (*model.User, error)
	Redirect(http.ResponseWriter, *http.Request)
	Login(*http.Request) (string, error)
	OnLoginError(http.ResponseWriter, *http.Request, error)
}
