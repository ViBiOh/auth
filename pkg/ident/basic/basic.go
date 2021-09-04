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

var _ ident.Provider = &App{}

// Provider check user credentials
type Provider interface {
	// Login user with its credentials
	Login(ctx context.Context, login, password string) (model.User, error)
}

// App of the package
type App struct {
	provider Provider
	realm    string
}

// New creates new App from Config
func New(provider Provider, realm string) App {
	return App{
		provider: provider,
		realm:    realm,
	}
}

// IsMatching checks if header content match provider
func (a App) IsMatching(content string) bool {
	return strings.HasPrefix(content, authPrefix)
}

// GetUser returns User found in content header
func (a App) GetUser(ctx context.Context, content string) (model.User, error) {
	rawData, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(content, authPrefix))
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

	return a.provider.Login(ctx, login, password)
}

// OnError handles HTTP request when login fails
func (a App) OnError(w http.ResponseWriter, _ *http.Request, err error) {
	realm := ""
	if len(a.realm) != 0 {
		realm = fmt.Sprintf("realm=\"%s\" ", a.realm)
	}

	w.Header().Add("WWW-Authenticate", fmt.Sprintf("Basic %scharset=\"UTF-8\"", realm))
	httperror.Unauthorized(w, err)
}
