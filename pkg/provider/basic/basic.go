package basic

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ViBiOh/auth/v2/pkg/model"
)

const (
	authPrefix = "Basic "
	lenPrefix  = len(authPrefix)
)

var (
	_ model.Identification = Service{}
	_ model.Authorization  = Service{}
)

type LoginProvider interface {
	Login(ctx context.Context, r *http.Request, login, password string) (model.User, error)
}

type PasswordStorage interface {
	SavePassword(ctx context.Context, user model.User, password string) error
	UpdatePassword(ctx context.Context, user model.User, password string) error
}

type Service struct {
	provider LoginProvider
	realm    string
}

func New(provider LoginProvider, realm string) Service {
	if len(realm) != 0 {
		realm = fmt.Sprintf("realm=\"%s\" ", realm)
	}

	return Service{
		provider: provider,
		realm:    realm,
	}
}
