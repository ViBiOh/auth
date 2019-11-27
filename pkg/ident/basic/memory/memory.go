package memory

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/ViBiOh/auth/pkg/ident"
	"github.com/ViBiOh/auth/pkg/ident/basic"
	"github.com/ViBiOh/auth/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/flags"
	"golang.org/x/crypto/bcrypt"
)

var _ basic.UserLogin = &app{}

type basicUser struct {
	model.User
	password []byte
}

// App of package
type App interface {
	Login(string, string) (model.User, error)
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
		users: flags.New(prefix, "basic").Name("Users").Default("").Label("Users in the form `id:login:password,id2:login2:password2`").ToString(fs),
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

func (a app) Login(login, password string) (model.User, error) {
	user, ok := a.users[login]
	if !ok {
		return model.NoneUser, ident.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword(user.password, []byte(password)); err != nil {
		return model.NoneUser, ident.ErrInvalidCredentials
	}

	return user.User, nil
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
				ID:    userID,
				Login: strings.ToLower(parts[1]),
			},
			password: []byte(parts[2]),
		}
		users[strings.ToLower(user.Login)] = user
	}

	return users, nil
}
