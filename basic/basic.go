package basic

import (
	"encoding/base64"
	"flag"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// User of the app
type User struct {
	Username string
	password []byte
}

var users map[string]*User

var (
	authUsers = flag.String(`basicUsers`, ``, `Basic users in the form "username:password,username2:password"`)
)

// Init auth
func Init() error {
	if err := LoadUsers(*authUsers); err != nil {
		return err
	}

	return nil
}

// LoadUsers loads given users into users map
func LoadUsers(authUsers string) error {
	users = make(map[string]*User)

	if authUsers == `` {
		return nil
	}

	for _, authUser := range strings.Split(authUsers, `,`) {
		parts := strings.Split(authUser, `:`)
		if len(parts) != 2 {
			return fmt.Errorf(`Invalid format of user for %s`, authUser)
		}

		user := User{strings.ToLower(parts[0]), []byte(parts[1])}
		users[strings.ToLower(user.Username)] = &user
	}

	return nil
}

// GetUsername returns username of given auth
func GetUsername(header string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(header)
	if err != nil {
		return ``, fmt.Errorf(`Error while decoding basic authentication: %v`, err)
	}

	dataStr := string(data)

	sepIndex := strings.Index(dataStr, `:`)
	if sepIndex < 0 {
		return ``, fmt.Errorf(`Error while reading basic authentication`)
	}

	username := dataStr[:sepIndex]
	password := dataStr[sepIndex+1:]

	if user, ok := users[strings.ToLower(username)]; ok {
		if err := bcrypt.CompareHashAndPassword(user.password, []byte(password)); err != nil {
			return ``, fmt.Errorf(`Invalid credentials for %s`, username)
		}
	}

	return username, nil
}
