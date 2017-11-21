package auth

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/auth/cookie"

	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/httputils/rate"
	"github.com/ViBiOh/httputils/tools"
)

const forbiddenMessage = `Not allowed to use app`
const authorizationHeader = `Authorization`
const forwardedForHeader = `X-Forwarded-For`

// ErrEmptyAuthorization occurs when authorization content is not found
var ErrEmptyAuthorization = errors.New(`Empty authorization content`)

// User of the app
type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	profiles string
}

// HasProfile check if User has given profile
func (user *User) HasProfile(profile string) bool {
	return strings.Contains(user.profiles, profile)
}

// Flags add flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`url`:   flag.String(tools.ToCamel(prefix+`Url`), ``, `[auth] Auth URL`),
		`users`: flag.String(tools.ToCamel(prefix+`Users`), ``, `[auth] List of allowed users and profiles (e.g. user:profile1|profile2,user2:profile3)`),
	}
}

// NewUser creates new user with given id, username and profiles
func NewUser(id uint, username string, profiles string) *User {
	return &User{ID: id, Username: username, profiles: profiles}
}

// LoadUsersProfiles parses users ands profiles from given string
func LoadUsersProfiles(usersAndProfiles string) map[string]*User {
	if usersAndProfiles == `` {
		return nil
	}

	users := make(map[string]*User, 0)

	for _, user := range strings.Split(usersAndProfiles, `,`) {
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

// IsForbiddenErr check if given error refer to a forbidden
func IsForbiddenErr(err error) bool {
	return strings.HasSuffix(err.Error(), forbiddenMessage)
}

// IsAuthenticated check if request has correct headers for authentification
func IsAuthenticated(url string, users map[string]*User, r *http.Request) (*User, error) {
	authorization := r.Header.Get(authorizationHeader)

	if authorization == `` {
		authorization, _ = cookie.GetCookieValue(r, `auth`)
	}

	return IsAuthenticatedByAuth(url, users, authorization, rate.GetIP(r))
}

// IsAuthenticatedByAuth check if authorization is correct
func IsAuthenticatedByAuth(url string, users map[string]*User, authContent, remoteIP string) (*User, error) {
	headers := map[string]string{
		authorizationHeader: authContent,
		forwardedForHeader:  remoteIP,
	}

	userBytes, err := httputils.GetBody(url+`/user`, headers)
	if err != nil {
		if strings.HasPrefix(string(userBytes), ErrEmptyAuthorization.Error()) {
			return nil, ErrEmptyAuthorization
		}
		return nil, fmt.Errorf(`Error while getting user: %v`, err)
	}

	user := User{}
	if err := json.Unmarshal(userBytes, &user); err != nil {
		return nil, fmt.Errorf(`Error while unmarshalling user: %v`, err)
	}

	if appUser, ok := users[strings.ToLower(string(user.Username))]; ok {
		appUser.ID = user.ID
		return appUser, nil
	}

	return nil, fmt.Errorf(`[%s] %s`, user.Username, forbiddenMessage)
}

// Handler wrap next authenticated handler
func Handler(url string, users map[string]*User, next func(http.ResponseWriter, *http.Request, *User)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if user, err := IsAuthenticated(url, users, r); err != nil {
			if IsForbiddenErr(err) {
				httputils.Forbidden(w)
			} else {
				httputils.Unauthorized(w, err)
			}
		} else {
			next(w, r, user)
		}
	})
}
