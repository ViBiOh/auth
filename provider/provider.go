package provider

import (
	"errors"
	"net/http"

	"github.com/ViBiOh/auth/auth"
)

// ErrUnknownAuthType occurs when authentification type in unknown
var ErrUnknownAuthType = errors.New(`Unknown authentication type`)

// ErrInvalidState occurs when state is not consistent
var ErrInvalidState = errors.New(`Invalid state provided for oauth`)

// ErrInvalidCode occurs when code is no valid
var ErrInvalidCode = errors.New(`Invalid code provided for oauth`)

// Auth is a provider of Authentification methods
type Auth interface {
	GetName() string
	GetUser(string) (*auth.User, error)
	Redirect() (string, error)
	Login(*http.Request) (string, error)
}
