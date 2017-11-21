package basic

import (
	"encoding/base64"
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/ViBiOh/auth/auth"
	"github.com/ViBiOh/auth/provider"
	"github.com/ViBiOh/httputils/tools"
	"golang.org/x/crypto/bcrypt"
)

type basicUser struct {
	*auth.User
	password []byte
}

// Flags add flags for given prefix
func Flags(prefix string) map[string]interface{} {
	return map[string]interface{}{
		`users`: flag.String(tools.ToCamel(prefix+`Users`), ``, `[Basic] Users in the form "id:username:password,id2:username2:password2"`),
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
			return nil, fmt.Errorf(`Invalid format of user for %s`, authUser)
		}

		id, err := strconv.ParseUint(parts[0], 10, 32)
		if err != nil {
			return nil, fmt.Errorf(`Invalid id format for user %s`, authUser)
		}

		user := basicUser{&auth.User{ID: uint(id), Username: strings.ToLower(parts[1])}, []byte(parts[2])}
		users[strings.ToLower(user.Username)] = &user
	}

	return users, nil
}

// Auth with login/pass
type Auth struct {
	users map[string]*basicUser
}

// Init provider
func (o *Auth) Init(config map[string]interface{}) error {
	users, err := loadUsers(*(config[`users`].(*string)))
	if err != nil {
		return err
	}

	o.users = users

	return nil
}

// GetName returns Authorization header prefix
func (*Auth) GetName() string {
	return `Basic`
}

// GetUser returns User associated to header
func (o *Auth) GetUser(header string) (*auth.User, error) {
	data, err := base64.StdEncoding.DecodeString(header)
	if err != nil {
		return nil, fmt.Errorf(`Error while decoding basic authentication: %v`, err)
	}

	dataStr := string(data)

	sepIndex := strings.Index(dataStr, `:`)
	if sepIndex < 0 {
		return nil, fmt.Errorf(`Error while reading basic authentication`)
	}

	username := strings.ToLower(dataStr[:sepIndex])
	password := dataStr[sepIndex+1:]

	user, ok := o.users[username]
	if ok {
		if err := bcrypt.CompareHashAndPassword(user.password, []byte(password)); err != nil {
			ok = false
		}
	}

	if !ok {
		return nil, fmt.Errorf(`Invalid credentials for %s`, username)
	}

	return user.User, nil
}

// Authorize redirect user to authorize endpoint
func (*Auth) Authorize() (string, map[string]string, error) {
	return ``, map[string]string{`WWW-Authenticate`: `Basic`}, nil
}

// GetAccessToken exchange state to token
func (*Auth) GetAccessToken(string, string, string) (string, error) {
	return ``, provider.ErrNoToken
}
