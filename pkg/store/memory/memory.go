package memory

import (
	"flag"
	"fmt"
	"strings"

	"github.com/ViBiOh/auth/v3/pkg/model"
	"github.com/ViBiOh/auth/v3/pkg/provider/basic"
	"github.com/ViBiOh/flags"
)

var (
	_ model.Storage  = Service{}
	_ basic.Provider = Service{}
)

type Service struct {
	identifications map[string]basicUser
	authorizations  map[string][]string
}

type Config struct {
	Idents []string
	Auths  []string
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) *Config {
	var config Config

	flags.New("Users", "Users credentials in the form 'id:login:password'").Prefix(prefix).DocPrefix("memory").EnvSeparator("|").StringSliceVar(fs, &config.Idents, nil, overrides)
	flags.New("Profiles", "Users profiles in the form 'id:profile1|profile2'").Prefix(prefix).DocPrefix("memory").StringSliceVar(fs, &config.Auths, nil, overrides)

	return &config
}

func New(config *Config) (Service, error) {
	identifications, err := loadIdent(config.Idents)
	if err != nil {
		return Service{}, fmt.Errorf("load ident: %w", err)
	}

	authorizations, err := loadAuth(config.Auths)
	if err != nil {
		return Service{}, fmt.Errorf("load auth: %w", err)
	}

	return Service{
		identifications: identifications,
		authorizations:  authorizations,
	}, nil
}

func loadIdent(idents []string) (map[string]basicUser, error) {
	if len(idents) == 0 {
		return nil, nil
	}

	users := make(map[string]basicUser)
	ids := make(map[string]struct{})

	for _, identUser := range idents {
		parts := strings.Split(identUser, ":")
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid format for user ident `%s`", identUser)
		}

		userID := strings.ToLower(parts[0])

		if _, ok := ids[userID]; ok {
			return nil, fmt.Errorf("id already exists for user ident `%s`", identUser)
		}
		ids[userID] = struct{}{}

		user := basicUser{
			User:     model.User{ID: userID, Name: strings.ToLower(parts[1])},
			password: []byte(parts[2]),
		}
		users[user.Name] = user
	}

	return users, nil
}

func loadAuth(auths []string) (map[string][]string, error) {
	if len(auths) == 0 {
		return nil, nil
	}

	users := make(map[string][]string)

	for _, authUser := range auths {
		parts := strings.Split(authUser, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid format of user auth `%s`", authUser)
		}

		users[parts[0]] = strings.Split(parts[1], "|")
	}

	return users, nil
}
