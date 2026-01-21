package model

import (
	"context"
	"errors"
	"net/http"
)

var (
	ErrForbidden          = errors.New("forbidden access")
	ErrUnknownUser        = errors.New("unknown user")
	ErrMalformedContent   = errors.New("malformed content")
	ErrUnavailableService = errors.New("unavailable service")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type Storage interface {
	Create(context.Context, string) (User, error)
	Delete(context.Context, User) error
}

type Authentication interface {
	GetUser(context.Context, http.ResponseWriter, *http.Request) (User, error)
	OnUnauthorized(http.ResponseWriter, *http.Request, error)
}

type Authorization interface {
	IsAuthorized(context.Context, *http.Request, User) bool
	OnForbidden(http.ResponseWriter, *http.Request, User)
}
