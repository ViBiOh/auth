package basic

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/auth/v2/pkg/ident"
	"github.com/ViBiOh/auth/v2/pkg/middleware"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/httperror"
)

const (
	authPrefix = "Basic "
	lenPrefix  = len(authPrefix)
)

var _ ident.Provider = Service{}

type LoginProvider interface {
	Login(ctx context.Context, login, password string) (model.User, error)
}

type Service struct {
	provider LoginProvider
	realm    string
}

func New(provider LoginProvider, realm string) Service {
	return Service{
		provider: provider,
		realm:    realm,
	}
}

func (s Service) GetUser(ctx context.Context, r *http.Request) (model.User, error) {
	content := r.Header.Get("Authorization")
	if len(content) == 0 || content[:lenPrefix] != authPrefix {
		return model.User{}, middleware.ErrEmptyAuth
	}

	if len(content) < lenPrefix {
		return model.User{}, ident.ErrMalformedAuth
	}

	rawData, err := base64.StdEncoding.DecodeString(content[lenPrefix:])
	if err != nil {
		return model.User{}, fmt.Errorf("%s: %w", err, ident.ErrMalformedAuth)
	}

	data := string(rawData)

	sepIndex := strings.Index(data, ":")
	if sepIndex == -1 {
		return model.User{}, ident.ErrMalformedAuth
	}

	login := strings.ToLower(data[:sepIndex])
	password := strings.TrimSuffix(data[sepIndex+1:], "\n")

	return s.provider.Login(ctx, login, password)
}

func (s Service) OnError(w http.ResponseWriter, r *http.Request, err error) {
	realm := ""
	if len(s.realm) != 0 {
		realm = fmt.Sprintf("realm=\"%s\" ", s.realm)
	}

	if errors.Is(err, middleware.ErrEmptyAuth) {
		err = nil // We don't want to log it
	}

	w.Header().Add("WWW-Authenticate", fmt.Sprintf("Basic %scharset=\"UTF-8\"", realm))
	httperror.Unauthorized(r.Context(), w, err)
}
