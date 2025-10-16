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

//go:generate mockgen -source=$GOFILE -destination=../mocks/$GOFILE -package=mocks -mock_names UpdatableStorage=UpdatableStorage

type Storage interface {
	Create(context.Context) (User, error)
	Delete(context.Context, User) error
}

type Identification interface {
	GetUser(context.Context, *http.Request) (User, error)
	OnError(http.ResponseWriter, *http.Request, error)
}

type Authorization interface {
	IsAuthorized(context.Context, User, string) bool
	OnForbidden(http.ResponseWriter, *http.Request, User, string)
}
