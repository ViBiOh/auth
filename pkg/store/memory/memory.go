package memory

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/ViBiOh/auth/v2/pkg/auth"
	"github.com/ViBiOh/auth/v2/pkg/ident/basic"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/flags"
)

var (
	_ auth.Provider  = Service{}
	_ basic.Provider = Service{}
)

type Service struct {
	ident map[string]basicUser
	auth  map[uint64][]string
}

type Config struct {
	Ident string
	Auth  string
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	var config Config

	flags.New("Users", "Users credentials in the form 'id:login:password,id2:login2:password2'").Prefix(prefix).DocPrefix("memory").StringVar(fs, &config.Ident, "", overrides)
	flags.New("Profiles", "Users profiles in the form 'id:profile1|profile2,id2:profile1'").Prefix(prefix).DocPrefix("memory").StringVar(fs, &config.Auth, "", overrides)

	return config
}

func New(config Config) (Service, error) {
	identService, err := loadIdent(config.Ident)
	if err != nil {
		return Service{}, fmt.Errorf("load ident: %w", err)
	}

	authService, err := loadAuth(config.Auth)
	if err != nil {
		return Service{}, fmt.Errorf("load auth: %w", err)
	}

	return Service{
		ident: identService,
		auth:  authService,
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
