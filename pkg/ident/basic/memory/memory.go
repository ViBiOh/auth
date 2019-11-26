package memory

import (
	"context"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ViBiOh/auth/pkg/ident"
	"github.com/ViBiOh/auth/pkg/model"
	"github.com/ViBiOh/httputils/pkg/httperror"
	"github.com/ViBiOh/httputils/v3/pkg/flags"
	"golang.org/x/crypto/bcrypt"
)

type basicUser struct {
	model.User
	password []byte
}

// App of package
type App interface {
	GetUser(context.Context, string) (model.User, error)
}

// Config of package
type Config struct {
	users *string
}

type app struct {
	users map[string]basicUser
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		users: flags.New(prefix, "basic").Name("Users").Default("").Label("Users in the form `id:username:password,id2:username2:password2`").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config) (App, error) {
	users, err := loadInMemoryUsers(strings.TrimSpace(*config.users))
	if err != nil {
		return nil, err
	}

	return &app{
		users: users,
	}, nil
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

	user, ok := a.users[username]
	if !ok {
		return model.NoneUser, ident.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword(user.password, []byte(password)); err != nil {
		return model.NoneUser, ident.ErrInvalidCredentials
	}

	return user.User, nil
}

// OnError handle action when login fails
func (a app) OnError(w http.ResponseWriter, _ *http.Request, err error) {
	w.Header().Add("WWW-Authenticate", "Basic charset=\"UTF-8\"")
	httperror.Unauthorized(w, err)
}

func loadInMemoryUsers(authUsers string) (map[string]basicUser, error) {
	if authUsers == "" {
		return nil, nil
	}

	users := make(map[string]basicUser)

	for _, authUser := range strings.Split(authUsers, ",") {
		parts := strings.Split(authUser, ":")
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid format of user for %s", authUser)
		}

		userID, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			return nil, err
		}

		user := basicUser{
			User: model.User{
				ID:       userID,
				Username: strings.ToLower(parts[1]),
			},
			password: []byte(parts[2]),
		}
		users[strings.ToLower(user.Username)] = user
	}

	return users, nil
}
