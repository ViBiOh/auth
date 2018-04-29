package provider

import (
	"errors"
	"net/http"

	"github.com/ViBiOh/auth/pkg/model"
)

var (
	// ErrUnknownAuthType occurs when authentification type in unknown
	ErrUnknownAuthType = errors.New(`Unknown authentication type`)

	// ErrInvalidState occurs when state is not consistent
	ErrInvalidState = errors.New(`Invalid state provided for oauth`)

	// ErrInvalidCode occurs when code is no valid
	ErrInvalidCode = errors.New(`Invalid code provided for oauth`)

	// ErrMalformedAuth occurs when auth header is malformed
	ErrMalformedAuth = errors.New(`Malformed Authorization content`)
)

// Service provide methods for dealing with authentification
type Service interface {
	GetUser(string) (*model.User, error)
}

// Auth is a provider of Authentification methods
type Auth interface {
	GetName() string
	GetUser(string) (*model.User, error)
	Redirect() (string, error)
	Login(*http.Request) (string, error)
}
