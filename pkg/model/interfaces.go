package model

import (
	"context"
	"errors"
	"net/http"
)

var (
	ErrForbidden          = errors.New("forbidden access")
	ErrMalformedContent   = errors.New("malformed content")
	ErrUnavailableService = errors.New("unavailable service")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

//go:generate mockgen -source=$GOFILE -destination=../mocks/$GOFILE -package=mocks -mock_names UpdatableStorage=UpdatableStorage

type Storage interface {
	Get(context.Context, uint64) (User, error)
}

type UpdatableStorage interface {
	Storage
	Create(context.Context, User) (uint64, error)
	Update(context.Context, User) error
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
