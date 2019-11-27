package memory

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/ViBiOh/auth/pkg/auth"
	"github.com/ViBiOh/auth/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/flags"
)

var _ auth.Provider = &App{}

// Config of package
type Config struct {
	users *string
}

// App of package
type App struct {
	users map[uint64][]string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		users: flags.New(prefix, "auth").Name("Users").Default("").Label("Users profiles in the form `id:profile1|profile2,id2:profile1`").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config) (App, error) {
	users, err := loadInMemoryUsers(strings.TrimSpace(*config.users))
	if err != nil {
		return App{}, err
	}

	return App{
		users: users,
	}, nil
}

// IsAuthorized checks if User is authorized
func (a App) IsAuthorized(user model.User, profile string) bool {
	profiles, ok := a.users[user.ID]
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

func loadInMemoryUsers(authUsers string) (map[uint64][]string, error) {
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
