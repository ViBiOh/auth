package basic

import (
	"encoding/base64"
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/ViBiOh/auth/auth"
	"golang.org/x/crypto/bcrypt"
)

type basicUser struct {
	*auth.User
	password []byte
}

var users map[string]*basicUser

var (
	authUsers = flag.String(`basicUsers`, ``, `Basic users in the form "id:username:password,id2:username2:password2"`)
)

// Init auth
func Init() error {
	return LoadUsers(*authUsers)
}

// LoadUsers loads given users into users map
func LoadUsers(authUsers string) error {
	users = make(map[string]*basicUser)

	if authUsers == `` {
		return nil
	}

	for _, authUser := range strings.Split(authUsers, `,`) {
		parts := strings.Split(authUser, `:`)
		if len(parts) != 3 {
			return fmt.Errorf(`Invalid format of user for %s`, authUser)
		}

		id, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return fmt.Errorf(`Invalid id format for user %s`, authUser)
		}

		user := basicUser{&auth.User{ID: id, Username: strings.ToLower(parts[1])}, []byte(parts[2])}
		users[strings.ToLower(user.Username)] = &user
	}

	return nil
}

// GetUser returns username of given auth
func GetUser(header string) (*auth.User, error) {
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

	user, ok := users[username]
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
