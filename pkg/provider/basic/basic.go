package basic

import (
	"context"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ViBiOh/auth/pkg/model"
	"github.com/ViBiOh/auth/pkg/provider"
	"github.com/ViBiOh/httputils/pkg/tools"
	"golang.org/x/crypto/bcrypt"
)

type basicUser struct {
	*model.User
	password []byte
}

// Flags add flags for given prefix
func Flags(prefix string) map[string]interface{} {
	return map[string]interface{}{
		`users`: flag.String(tools.ToCamel(fmt.Sprintf(`%s%s`, prefix, `Users`)), ``, `[Basic] Users in the form "id:username:password,id2:username2:password2"`),
	}
}

func loadUsers(authUsers string) (map[string]*basicUser, error) {
	users := make(map[string]*basicUser)

	if authUsers == `` {
		return nil, nil
	}

	for _, authUser := range strings.Split(authUsers, `,`) {
		parts := strings.Split(authUser, `:`)
		if len(parts) != 3 {
			return nil, fmt.Errorf(`invalid format of user for %s`, authUser)
		}

		id, err := strconv.ParseUint(parts[0], 10, 32)
		if err != nil {
			return nil, fmt.Errorf(`invalid id format for user %s`, authUser)
		}

		user := basicUser{&model.User{ID: uint(id), Username: strings.ToLower(parts[1])}, []byte(parts[2])}
		users[strings.ToLower(user.Username)] = &user
	}

	return users, nil
}

// Auth with login/pass
type Auth struct {
	users map[string]*basicUser
}

// NewAuth creates new auth
func NewAuth(config map[string]interface{}) (provider.Auth, error) {
	users, err := loadUsers(*(config[`users`].(*string)))

	if err != nil {
		return nil, fmt.Errorf(`error while loading users: %v`, err)
	}

	if users != nil {
		return &Auth{users}, nil
	}

	return nil, nil
}

// GetName returns Authorization header prefix
func (*Auth) GetName() string {
	return `Basic`
}

// GetUser returns User associated to header
func (a *Auth) GetUser(ctx context.Context, header string) (*model.User, error) {
	data, err := base64.StdEncoding.DecodeString(header)
	if err != nil {
		return nil, fmt.Errorf(`error while decoding basic authentication: %v`, err)
	}

	dataStr := string(data)

	sepIndex := strings.Index(dataStr, `:`)
	if sepIndex < 0 {
		return nil, errors.New(`error while reading basic authentication`)
	}

	username := strings.ToLower(dataStr[:sepIndex])
	password := dataStr[sepIndex+1:]

	user, ok := a.users[username]
	if ok {
		if err := bcrypt.CompareHashAndPassword(user.password, []byte(password)); err != nil {
			ok = false
		}
	}

	if !ok {
		return nil, fmt.Errorf(`invalid credentials for %s`, username)
	}

	return user.User, nil
}

// Redirect redirects user to login endpoint
func (*Auth) Redirect() (string, error) {
	return `/login/basic`, nil
}

// Login exchange state to token
func (a *Auth) Login(r *http.Request) (string, error) {
	authContent := strings.TrimPrefix(r.Header.Get(`Authorization`), fmt.Sprintf(`%s `, a.GetName()))

	if _, err := a.GetUser(r.Context(), authContent); err != nil {
		return ``, err
	}
	return authContent, nil
}
