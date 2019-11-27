package basic

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/ViBiOh/auth/pkg/ident"
	"github.com/ViBiOh/auth/pkg/model"
	"github.com/ViBiOh/httputils/pkg/httperror"
)

const (
	authPrefix = "Basic "
)

var _ ident.Provider = &App{}

// UserLogin login user based on its credentials
type UserLogin interface {
	// Login user with its credentials
	Login(string, string) (model.User, error)
}

// App of the package
type App struct {
	userLogin UserLogin
}

// New creates new App from Config
func New(userLogin UserLogin) App {
	return App{
		userLogin: userLogin,
	}
}

// IsMatching checks if header content match provider
func (a App) IsMatching(content string) bool {
	return strings.HasPrefix(content, authPrefix)
}

// GetUser returns User found in content header
func (a App) GetUser(content string) (model.User, error) {
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
	password := data[sepIndex+1:]

	return a.userLogin.Login(login, password)
}

// OnError handles HTTP request when login fails
func (a App) OnError(w http.ResponseWriter, _ *http.Request, err error) {
	w.Header().Add("WWW-Authenticate", "Basic charset=\"UTF-8\"")
	httperror.Unauthorized(w, err)
}
