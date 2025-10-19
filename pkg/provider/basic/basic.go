package basic

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ViBiOh/auth/v3/pkg/model"
)

const (
	authPrefix = "Basic "
	lenPrefix  = len(authPrefix)
)

var (
	_ model.Identification = Service{}
	_ model.Authorization  = Service{}
)

type Provider interface {
	GetBasicUser(ctx context.Context, r *http.Request, login, password string) (model.User, error)
	IsAuthorized(ctx context.Context, user model.User, profile string) bool
}

type PasswordStorage interface {
	SavePassword(ctx context.Context, user model.User, password string) error
	UpdatePassword(ctx context.Context, user model.User, password string) error
}

type ForbiddenHandler func(http.ResponseWriter, *http.Request, model.User, string)

type Service struct {
	provider    Provider
	onForbidden ForbiddenHandler
	realm       string
}

func New(provider Provider, options ...Option) Service {
	service := Service{
		provider: provider,
	}

	for _, option := range options {
		service = option(service)
	}

	return service
}

type Option func(Service) Service

func WithRealm(realm string) Option {
	return func(instance Service) Service {
		if len(realm) != 0 {
			instance.realm = fmt.Sprintf("realm=\"%s\" ", realm)
		}

		return instance
	}
}

func WithForbiddenHandler(onForbidden ForbiddenHandler) Option {
	return func(instance Service) Service {
		instance.onForbidden = onForbidden

		return instance
	}
}
