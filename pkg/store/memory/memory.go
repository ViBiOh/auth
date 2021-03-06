package memory

import (
	"context"
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/ViBiOh/auth/v2/pkg/auth"
	"github.com/ViBiOh/auth/v2/pkg/ident/basic"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/flags"
)

var (
	_ auth.Provider  = app{}
	_ basic.Provider = app{}
)

// App of package
type App interface {
	Login(ctx context.Context, login, password string) (model.User, error)
	IsAuthorized(ctx context.Context, user model.User, profile string) bool
}

// Config of package
type Config struct {
	ident *string
	auth  *string
}

type app struct {
	ident map[string]basicUser
	auth  map[uint64][]string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		ident: flags.New(prefix, "memory").Name("Users").Default("").Label("Users credentials in the form 'id:login:password,id2:login2:password2'").ToString(fs),
		auth:  flags.New(prefix, "memory").Name("Profiles").Default("").Label("Users profiles in the form 'id:profile1|profile2,id2:profile1'").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config) (App, error) {
	identApp, err := loadIdent(strings.TrimSpace(*config.ident))
	if err != nil {
		return nil, err
	}

	authApp, err := loadAuth(strings.TrimSpace(*config.auth))
	if err != nil {
		return nil, err
	}

	return &app{
		ident: identApp,
		auth:  authApp,
	}, nil
}

func loadIdent(ident string) (map[string]basicUser, error) {
	if len(ident) == 0 {
		return nil, nil
	}

	users := make(map[string]basicUser)
	ids := make(map[uint64]bool)

	for _, identUser := range strings.Split(ident, ",") {
		parts := strings.Split(identUser, ":")
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid format for user ident `%s`", identUser)
		}

		userID, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			return nil, err
		}

		if ids[userID] {
			return nil, fmt.Errorf("id already exists for user ident `%s`", identUser)
		}
		ids[userID] = true

		user := basicUser{
			User:     model.NewUser(userID, strings.ToLower(parts[1])),
			password: []byte(parts[2]),
		}
		users[user.Login] = user
	}

	return users, nil
}

func loadAuth(auth string) (map[uint64][]string, error) {
	if len(auth) == 0 {
		return nil, nil
	}

	users := make(map[uint64][]string)

	for _, authUser := range strings.Split(auth, ",") {
		parts := strings.Split(authUser, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid format of user auth `%s`", authUser)
		}

		userID, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			return nil, err
		}

		users[userID] = strings.Split(parts[1], "|")
	}

	return users, nil
}
