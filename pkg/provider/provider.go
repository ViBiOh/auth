package provider

import (
	"errors"
	"net/http"
	"strings"
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

// User of the app
type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	profiles string
}

// NewUser creates new user with given id, username and profiles
func NewUser(id uint, username string, profiles string) *User {
	return &User{ID: id, Username: username, profiles: profiles}
}

// HasProfile check if User has given profile
func (user *User) HasProfile(profile string) bool {
	return strings.Contains(user.profiles, profile)
}

// Service provide methods for dealing with authentification
type Service interface {
	GetUser(string) (*User, error)
}

// Auth is a provider of Authentification methods
type Auth interface {
	GetName() string
	GetUser(string) (*User, error)
	Redirect() (string, error)
	Login(*http.Request) (string, error)
}
