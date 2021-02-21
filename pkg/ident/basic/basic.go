package basic

import (
	"context"
	"encoding/base64"
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
}

// New creates new App from Config
func New(provider Provider) App {
	return App{
		provider: provider,
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
		return model.NoneUser, ident.ErrMalformedAuth
	}

	data := string(rawData)

	sepIndex := strings.Index(data, ":")
	if sepIndex < 0 {
		return model.NoneUser, ident.ErrMalformedAuth
	}

	login := strings.ToLower(data[:sepIndex])
	password := strings.TrimSpace(data[sepIndex+1:])

	return a.provider.Login(ctx, login, password)
}

// OnError handles HTTP request when login fails
func (a App) OnError(w http.ResponseWriter, _ *http.Request, err error) {
	w.Header().Add("WWW-Authenticate", "Basic charset=\"UTF-8\"")
	httperror.Unauthorized(w, err)
}
