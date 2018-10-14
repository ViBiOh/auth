package auth

import (
	"context"
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
var ErrEmptyAuthorization = errors.New(`empty authorization content`)

// App stores informations and secret of API
type App struct {
	serviceApp provider.Service
	URL        string
	users      map[string]string
	disabled   bool
}

// NewApp creates new App from Flags' config
func NewApp(config map[string]interface{}, serviceApp provider.Service) *App {
	return &App{
		serviceApp: serviceApp,
		URL:        strings.TrimSpace(*config[`url`].(*string)),
		users:      loadUsersProfiles(*config[`users`].(*string)),
		disabled:   *config[`disable`].(*bool),
	}
}

// Flags add flags for given prefix
func Flags(prefix string) map[string]interface{} {
	return map[string]interface{}{
		`disable`: flag.Bool(tools.ToCamel(fmt.Sprintf(`%sDisable`, prefix)), false, `[auth] Disable auth`),
		`url`:     flag.String(tools.ToCamel(fmt.Sprintf(`%sUrl`, prefix)), ``, `[auth] Auth URL, if remote`),
		`users`:   flag.String(tools.ToCamel(fmt.Sprintf(`%sUsers`, prefix)), ``, `[auth] List of allowed users and profiles (e.g. user:profile1|profile2,user2:profile3)`),
	}
}

// IsAuthenticated check if request has correct headers for authentification
func (a App) IsAuthenticated(r *http.Request) (*model.User, error) {
	return a.IsAuthenticatedByAuth(r.Context(), ReadAuthContent(r))
}

// IsAuthenticatedByAuth check if authorization is correct
func (a App) IsAuthenticatedByAuth(ctx context.Context, authContent string) (*model.User, error) {
	var retrievedUser *model.User
	var err error

	if a.serviceApp == nil && a.URL == `` {
		return nil, errors.New(`no authentification target configured`)
	}

	if a.serviceApp != nil {
		retrievedUser, err = a.serviceApp.GetUser(ctx, authContent)
		if err != nil && a.URL == `` {
			return nil, fmt.Errorf(`error while getting user from service: %v`, err)
		}
	}

	if retrievedUser == nil && a.URL != `` {
		headers := http.Header{}
		headers.Set(authorizationHeader, authContent)

		userBytes, err := request.Get(ctx, fmt.Sprintf(`%s/user`, a.URL), headers)
		if err != nil {
			if strings.HasPrefix(string(userBytes), ErrEmptyAuthorization.Error()) {
				return nil, ErrEmptyAuthorization
			}

			return nil, fmt.Errorf(`error while getting user from remote: %v`, err)
		}

		retrievedUser = &model.User{}
		if err := json.Unmarshal(userBytes, retrievedUser); err != nil {
			return nil, fmt.Errorf(`error while unmarshalling user: %v`, err)
		}
	}

	username := strings.ToLower(retrievedUser.Username)
	if profiles, ok := a.users[username]; ok {
		return model.NewUser(retrievedUser.ID, username, retrievedUser.Email, profiles), nil
	}

	return nil, fmt.Errorf(`%s: %s`, retrievedUser.Username, forbiddenMessage)
}

// HandlerWithFail wrap next authenticated handler and fail handler
func (a App) HandlerWithFail(next func(http.ResponseWriter, *http.Request, *model.User), fail func(http.ResponseWriter, *http.Request, error)) http.Handler {
	if a.disabled {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next(w, r, nil)
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if user, err := a.IsAuthenticated(r); err != nil {
			fail(w, r, err)
		} else {
			next(w, r, user)
		}
	})
}

// Handler wrap next authenticated handler
func (a App) Handler(next func(http.ResponseWriter, *http.Request, *model.User)) http.Handler {
	if a.disabled {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next(w, r, nil)
		})
	}

	return a.HandlerWithFail(next, defaultFailFunc)
}

// IsForbiddenErr check if given error refer to a forbidden
func IsForbiddenErr(err error) bool {
	return strings.HasSuffix(err.Error(), forbiddenMessage)
}

func loadUsersProfiles(usersAndProfiles string) map[string]string {
	if usersAndProfiles == `` {
		return nil
	}

	users := make(map[string]string, 0)

	for _, user := range strings.Split(usersAndProfiles, `,`) {
		username := user
		profiles := ``

		if parts := strings.Split(user, `:`); len(parts) == 2 {
			username = parts[0]
			profiles = parts[1]
		}

		users[strings.ToLower(username)] = profiles
	}

	return users
}

// ReadAuthContent from Header or Cookie
func ReadAuthContent(r *http.Request) string {
	authContent := strings.TrimSpace(r.Header.Get(authorizationHeader))
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
