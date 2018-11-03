package auth

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/auth/pkg/cookie"
	"github.com/ViBiOh/auth/pkg/model"
	"github.com/ViBiOh/auth/pkg/provider"
	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/httputils/pkg/httperror"
	http_model "github.com/ViBiOh/httputils/pkg/model"
	"github.com/ViBiOh/httputils/pkg/request"
	"github.com/ViBiOh/httputils/pkg/tools"
)

type key int

const (
	authorizationHeader     = `Authorization`
	ctxUserName         key = iota
)

var _ http_model.Middleware = &App{}

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

// UserFromContext retrieves user from context
func UserFromContext(ctx context.Context) *model.User {
	rawUser := ctx.Value(ctxUserName)
	if rawUser == nil {
		return nil
	}

	if user, ok := rawUser.(*model.User); ok {
		return user
	}
	return nil
}

// ReadAuthContent from Header or Cookie
func ReadAuthContent(r *http.Request) string {
	authContent := strings.TrimSpace(r.Header.Get(authorizationHeader))
	if authContent != `` {
		return authContent
	}

	return cookie.GetCookieValue(r, `auth`)
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
			return nil, err
		}
	}

	if retrievedUser == nil && a.URL != `` {
		headers := http.Header{}
		headers.Set(authorizationHeader, authContent)

		userBytes, status, _, err := request.Get(ctx, fmt.Sprintf(`%s/user`, a.URL), headers)
		if err != nil {
			if status == http.StatusUnauthorized {
				return nil, provider.ErrEmptyAuthorization
			}

			return nil, errors.New(`authentication failed: %v`, err)
		}

		retrievedUser = &model.User{}
		if err := json.Unmarshal(userBytes, retrievedUser); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	username := strings.ToLower(retrievedUser.Username)
	if profiles, ok := a.users[username]; ok {
		return model.NewUser(retrievedUser.ID, username, retrievedUser.Email, profiles), nil
	}

	return nil, provider.ErrForbiden
}

// HandlerWithFail wrap next authenticated handler and fail handler
func (a App) HandlerWithFail(next http.Handler, fail func(http.ResponseWriter, *http.Request, error)) http.Handler {
	if a.disabled {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := a.IsAuthenticated(r)
		if err != nil {
			fail(w, r, err)
			return
		}

		ctx := context.WithValue(r.Context(), ctxUserName, user)
		r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// Handler wrap next authenticated handler
func (a App) Handler(next http.Handler) http.Handler {
	return a.HandlerWithFail(next, a.defaultFailFunc)
}

func (a App) defaultFailFunc(w http.ResponseWriter, r *http.Request, err error) {
	if err == provider.ErrEmptyAuthorization {
		if a.serviceApp != nil {
			a.serviceApp.OnError(w, r, err)
		}

		if a.URL != `` {
			if _, _, _, err := request.Get(r.Context(), fmt.Sprintf(`%s/redirect`, a.URL), nil); err != nil {
				httperror.InternalServerError(w, err)
				return
			}
		}
	}

	if err == provider.ErrForbiden {
		httperror.Forbidden(w)
		return
	}

	httperror.Unauthorized(w, err)
}
