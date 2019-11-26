package db

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"github.com/ViBiOh/auth/pkg/ident"
	"github.com/ViBiOh/auth/pkg/model"
	"github.com/ViBiOh/httputils/pkg/httperror"
	"github.com/ViBiOh/httputils/pkg/logger"
)

// App of package
type App interface {
	GetUser(context.Context, string) (model.User, error)
}

type app struct {
	db *sql.DB
}

// New creates new App from Config
func New(db *sql.DB) App {
	return &app{
		db: db,
	}
}

// GetUser returns User associated to header
func (a app) GetUser(ctx context.Context, header string) (model.User, error) {
	data, err := base64.StdEncoding.DecodeString(header)
	if err != nil {
		return model.NoneUser, err
	}

	dataStr := string(data)

	sepIndex := strings.Index(dataStr, ":")
	if sepIndex < 0 {
		return model.NoneUser, errors.New("invalid format for basic auth")
	}

	username := strings.ToLower(dataStr[:sepIndex])
	password := dataStr[sepIndex+1:]

	user, err := a.getUserFromCredentials(username, password)
	if err != nil {
		logger.Error("%s", err.Error())
		return model.NoneUser, ident.ErrUnavailableService
	}

	if user == model.NoneUser {
		return user, ident.ErrInvalidCredentials
	}

	return user, nil
}

// OnError handle action when login fails
func (a app) OnError(w http.ResponseWriter, _ *http.Request, err error) {
	w.Header().Add("WWW-Authenticate", "Basic charset=\"UTF-8\"")
	httperror.Unauthorized(w, err)
}
