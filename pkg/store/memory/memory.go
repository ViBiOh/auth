package memory

import (
	"context"
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/ViBiOh/auth/v2/pkg/auth"
	"github.com/ViBiOh/auth/v2/pkg/ident"
	"github.com/ViBiOh/auth/v2/pkg/ident/basic"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/flags"
	"golang.org/x/crypto/bcrypt"
)

var (
	_ auth.Provider  = &App{}
	_ basic.Provider = &App{}
)

type basicUser struct {
	model.User
	password []byte
}

// Config of package
type Config struct {
	basic *string
	auth  *string
}

// App of package
type App struct {
	basic map[string]basicUser
	auth  map[uint64][]string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		basic: flags.New(prefix, "memory").Name("Users").Default("").Label("Users credentials in the form `id:login:password,id2:login2:password2`").ToString(fs),
		auth:  flags.New(prefix, "memory").Name("Profiles").Default("").Label("Users profiles in the form `id:profile1|profile2,id2:profile1`").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config) (App, error) {
	basic, err := loadBasicUsers(strings.TrimSpace(*config.basic))
	if err != nil {
		return App{}, err
	}

	auth, err := loadAuthUsers(strings.TrimSpace(*config.auth))
	if err != nil {
		return App{}, err
	}

	return App{
		basic: basic,
		auth:  auth,
	}, nil
}

// Login user with its credentials
func (a App) Login(ctx context.Context, login, password string) (model.User, error) {
	user, ok := a.basic[login]
	if !ok {
		return model.NoneUser, ident.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword(user.password, []byte(password)); err != nil {
		return model.NoneUser, ident.ErrInvalidCredentials
	}

	return user.User, nil
}

// IsAuthorized checks if User is authorized
func (a App) IsAuthorized(ctx context.Context, user model.User, profile string) bool {
	profiles, ok := a.auth[user.ID]
	if !ok {
		return false
	}

	if len(strings.TrimSpace(profile)) == 0 {
		return true
	}

	for _, listedProfile := range profiles {
		if strings.EqualFold(profile, listedProfile) {
			return true
		}
	}

	return false
}

func loadBasicUsers(identUsers string) (map[string]basicUser, error) {
	if identUsers == "" {
		return nil, nil
	}

	users := make(map[string]basicUser)

	for _, identUser := range strings.Split(identUsers, ",") {
		parts := strings.Split(identUser, ":")
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid format for user login `%s`", identUser)
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

func loadAuthUsers(authUsers string) (map[uint64][]string, error) {
	if authUsers == "" {
		return nil, nil
	}

	users := make(map[uint64][]string)

	for _, authUser := range strings.Split(authUsers, ",") {
		parts := strings.Split(authUser, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid format of user profile `%s`", authUser)
		}

		userID, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			return nil, err
		}

		users[userID] = strings.Split(parts[1], "|")
	}

	return users, nil
}
