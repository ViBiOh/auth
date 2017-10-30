package auth

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/httputils/rate"
	"github.com/ViBiOh/httputils/tools"
)

const authorizationHeader = `Authorization`
const forwardedForHeader = `X-Forwarded-For`

// User of the app
type User struct {
	Username string
	profiles string
}

// HasProfile check if given User has given profile
func (user *User) HasProfile(profile string) bool {
	return strings.Contains(user.profiles, profile)
}

// Flags add flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`url`:   flag.String(tools.ToCamel(prefix+`Url`), ``, `[`+prefix+`] Auth URL`),
		`users`: flag.String(tools.ToCamel(prefix+`Users`), ``, `[`+prefix+`] List of allowed users and profiles (e.g. user:profile1,profile2|user2:profile3)`),
	}
}

// NewUser creates new user with username and profiles
func NewUser(username string, profiles string) *User {
	return &User{username, profiles}
}

// LoadUsersProfiles parses users ands profiles from given string
func LoadUsersProfiles(usersAndProfiles string) map[string]*User {
	users := make(map[string]*User, 0)

	if usersAndProfiles == `` {
		return nil
	}

	usersList := strings.Split(usersAndProfiles, `|`)
	for _, user := range usersList {
		username := user
		profiles := ``

		sepIndex := strings.Index(user, `:`)
		if sepIndex != -1 {
			username = user[:sepIndex]
			profiles = user[sepIndex+1:]
		}

		users[strings.ToLower(username)] = NewUser(username, profiles)
	}

	return users
}

// IsAuthenticated check if request has correct headers for authentification
func IsAuthenticated(url string, users map[string]*User, r *http.Request) (*User, error) {
	return IsAuthenticatedByAuth(url, users, r.Header.Get(authorizationHeader), rate.GetIP(r))
}

// IsAuthenticatedByAuth check if authorization is correct
func IsAuthenticatedByAuth(url string, users map[string]*User, authContent, remoteIP string) (*User, error) {
	headers := map[string]string{
		authorizationHeader: authContent,
		forwardedForHeader:  remoteIP,
	}

	username, err := httputils.GetBody(url+`/user`, headers, true)
	if err != nil {
		return nil, fmt.Errorf(`Error while getting username: %v`, err)
	}

	if user, ok := users[strings.ToLower(string(username))]; ok {
		return user, nil
	}

	return nil, fmt.Errorf(`[%s] Not allowed to use app`, username)
}