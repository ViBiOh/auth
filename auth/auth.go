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
	ID       int64  `json:"id"`
	Username string `json:"username"`
	profiles string
}

// HasProfile check if given User has given profile
func (user *User) HasProfile(profile string) bool {
	return strings.Contains(user.profiles, profile)
}

// Flags add flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`url`:   flag.String(tools.ToCamel(prefix+`Url`), ``, `[auth] Auth URL`),
		`users`: flag.String(tools.ToCamel(prefix+`Users`), ``, `[auth] List of allowed users and profiles (e.g. user:profile1,profile2|user2:profile3)`),
	}
}

// NewUser creates new user with username and profiles
func NewUser(id int64, username string, profiles string) *User {
	return &User{ID: id, Username: username, profiles: profiles}
}

// LoadUsersProfiles parses users ands profiles from given string
func LoadUsersProfiles(usersAndProfiles string) map[string]*User {
	if usersAndProfiles == `` {
		return nil
	}

	users := make(map[string]*User, 0)

	for _, user := range strings.Split(usersAndProfiles, `|`) {
		username := user
		profiles := ``

		sepIndex := strings.Index(user, `:`)
		if sepIndex != -1 {
			username = user[:sepIndex]
			profiles = user[sepIndex+1:]
		}

		users[strings.ToLower(username)] = NewUser(0, username, profiles)
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
