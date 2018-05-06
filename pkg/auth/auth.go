package auth

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/auth/pkg/cookie"
	"github.com/ViBiOh/auth/pkg/model"
	"github.com/ViBiOh/auth/pkg/provider"
	"github.com/ViBiOh/httputils/pkg/httperror"
	"github.com/ViBiOh/httputils/pkg/request"
	"github.com/ViBiOh/httputils/pkg/tools"
)

const (
	forbiddenMessage    = `Not allowed to use app`
	authorizationHeader = `Authorization`
)

// ErrEmptyAuthorization occurs when authorization content is not found
var ErrEmptyAuthorization = errors.New(`Empty authorization content`)

// App stores informations and secret of API
type App struct {
	URL        string
	serviceApp provider.Service
	users      map[string]*model.User
}

// NewApp creates new App from Flags' config
func NewApp(config map[string]*string, serviceApp provider.Service) *App {
	return &App{
		URL:        *config[`url`],
		serviceApp: serviceApp,
		users:      loadUsersProfiles(*config[`users`]),
	}
}

// Flags add flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`url`:   flag.String(tools.ToCamel(fmt.Sprintf(`%sUrl`, prefix)), ``, `[auth] Auth URL, if remote`),
		`users`: flag.String(tools.ToCamel(fmt.Sprintf(`%sUsers`, prefix)), ``, `[auth] List of allowed users and profiles (e.g. user:profile1|profile2,user2:profile3)`),
	}
}

// IsAuthenticated check if request has correct headers for authentification
func (a *App) IsAuthenticated(r *http.Request) (*model.User, error) {
	return a.IsAuthenticatedByAuth(ReadAuthContent(r))
}

// IsAuthenticatedByAuth check if authorization is correct
func (a *App) IsAuthenticatedByAuth(authContent string) (*model.User, error) {
	var retrievedUser *model.User
	var err error

	if a.serviceApp == nil && a.URL == `` {
		return nil, errors.New(`No authentification target configured`)
	}

	if a.serviceApp != nil {
		retrievedUser, err = a.serviceApp.GetUser(authContent)
		if err != nil && a.URL == `` {
			return nil, fmt.Errorf(`Error while getting user from service: %v`, err)
		}
	}

	if retrievedUser == nil && a.URL != `` {
		headers := map[string]string{
			authorizationHeader: authContent,
		}

		userBytes, err := request.Get(fmt.Sprintf(`%s/user`, a.URL), headers)
		if err != nil {
			if strings.HasPrefix(string(userBytes), ErrEmptyAuthorization.Error()) {
				return nil, ErrEmptyAuthorization
			}

			return nil, fmt.Errorf(`Error while getting user from remote: %v`, err)
		}

		retrievedUser = &model.User{}
		if err := json.Unmarshal(userBytes, retrievedUser); err != nil {
			return nil, fmt.Errorf(`Error while unmarshalling user: %v`, err)
		}
	}

	if appUser, ok := a.users[strings.ToLower(string(retrievedUser.Username))]; ok {
		appUser.ID = retrievedUser.ID
		return appUser, nil
	}

	return nil, fmt.Errorf(`[%s] %s`, retrievedUser.Username, forbiddenMessage)
}

// HandlerWithFail wrap next authenticated handler and fail handler
func (a *App) HandlerWithFail(next func(http.ResponseWriter, *http.Request, *model.User), fail func(http.ResponseWriter, *http.Request, error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if user, err := a.IsAuthenticated(r); err != nil {
			fail(w, r, err)
		} else {
			next(w, r, user)
		}
	})
}

// Handler wrap next authenticated handler
func (a *App) Handler(next func(http.ResponseWriter, *http.Request, *model.User)) http.Handler {
	return a.HandlerWithFail(next, defaultFailFunc)
}

// IsForbiddenErr check if given error refer to a forbidden
func IsForbiddenErr(err error) bool {
	return strings.HasSuffix(err.Error(), forbiddenMessage)
}

func loadUsersProfiles(usersAndProfiles string) map[string]*model.User {
	if usersAndProfiles == `` {
		return nil
	}

	users := make(map[string]*model.User, 0)

	for _, user := range strings.Split(usersAndProfiles, `,`) {
		username := user
		profiles := ``

		if parts := strings.Split(user, `:`); len(parts) == 2 {
			username = parts[0]
			profiles = parts[1]
		}

		users[strings.ToLower(username)] = model.NewUser(uint(len(users)), username, profiles)
	}

	return users
}

// ReadAuthContent from Header or Cookie
func ReadAuthContent(r *http.Request) string {
	authContent := r.Header.Get(authorizationHeader)
	if authContent != `` {
		return authContent
	}

	return cookie.GetCookieValue(r, `auth`)
}

func defaultFailFunc(w http.ResponseWriter, r *http.Request, err error) {
	if IsForbiddenErr(err) {
		httperror.Forbidden(w)
	} else {
		httperror.Unauthorized(w, err)
	}
}
