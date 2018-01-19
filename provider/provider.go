package provider

import (
	"errors"
	"net/http"

	"github.com/ViBiOh/auth/auth"
)

var (
	// ErrUnknownAuthType occurs when authentification type in unknown
	ErrUnknownAuthType = errors.New(`Unknown authentication type`)

	// ErrInvalidState occurs when state is not consistent
	ErrInvalidState = errors.New(`Invalid state provided for oauth`)

	// ErrInvalidCode occurs when code is no valid
	ErrInvalidCode = errors.New(`Invalid code provided for oauth`)
)

// Auth is a provider of Authentification methods
type Auth interface {
	GetName() string
	GetUser(string) (*auth.User, error)
	Redirect() (string, error)
	Login(*http.Request) (string, error)
}
