package basic

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/auth/v2/pkg/ident"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/httperror"
)

const (
	authPrefix = "Basic "
)

var _ ident.Provider = &Service{}

type Provider interface {
	Login(ctx context.Context, login, password string) (model.User, error)
}

type Service struct {
	provider Provider
	realm    string
}

func New(provider Provider, realm string) Service {
	return Service{
		provider: provider,
		realm:    realm,
	}
}

func (s Service) IsMatching(content string) bool {
	if len(content) < len(authPrefix) {
		return false
	}

	return content[:len(authPrefix)] == authPrefix
}

func (s Service) GetUser(ctx context.Context, content string) (model.User, error) {
	if len(content) < len(authPrefix) {
		return model.User{}, ident.ErrMalformedAuth
	}

	rawData, err := base64.StdEncoding.DecodeString(content[len(authPrefix):])
	if err != nil {
		return model.User{}, ident.ErrMalformedAuth
	}

	data := string(rawData)

	sepIndex := strings.Index(data, ":")
	if sepIndex < 0 {
		return model.User{}, ident.ErrMalformedAuth
	}

	login := strings.ToLower(data[:sepIndex])
	password := strings.TrimSuffix(data[sepIndex+1:], "\n")

	return s.provider.Login(ctx, login, password)
}

func (s Service) OnError(w http.ResponseWriter, _ *http.Request, err error) {
	realm := ""
	if len(s.realm) != 0 {
		realm = fmt.Sprintf("realm=\"%s\" ", s.realm)
	}

	w.Header().Add("WWW-Authenticate", fmt.Sprintf("Basic %scharset=\"UTF-8\"", realm))
	httperror.Unauthorized(w, err)
}
