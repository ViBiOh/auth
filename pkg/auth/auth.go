package auth

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/auth/pkg/cookie"
	"github.com/ViBiOh/auth/pkg/ident"
	"github.com/ViBiOh/auth/pkg/model"
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

var (
	// ErrForbidden occurs when user is identified but not authorized
	ErrForbidden = errors.New(`forbidden access`)

	_ http_model.Middleware = &App{}
)

// App stores informations and secret of API
type App struct {
	identService ident.Service
	URL          string
	users        map[string]string
	disabled     bool
}

// NewApp creates new App from Flags' config wuth url
func NewApp(config map[string]interface{}) *App {
	return &App{
		URL:      strings.TrimSpace(*config[`url`].(*string)),
		users:    loadUsersProfiles(*config[`users`].(*string)),
		disabled: *config[`disable`].(*bool),
	}
}

// NewServiceApp creates new App from Flags' config with service
func NewServiceApp(config map[string]interface{}, identService ident.Service) *App {
	return &App{
		identService: identService,
		disabled:     *config[`disable`].(*bool),
	}
}

// Flags add flags for given prefix
func Flags(prefix string) map[string]interface{} {
	return map[string]interface{}{
		`disable`: flag.Bool(tools.ToCamel(fmt.Sprintf(`%sDisable`, prefix)), false, `[auth] Disable auth`),
		`url`:     flag.String(tools.ToCamel(fmt.Sprintf(`%sUrl`, prefix)), ``, `[auth] Auth URL, if remote`),
		`users`:   flag.String(tools.ToCamel(fmt.Sprintf(`%sUsers`, prefix)), ``, `[auth] Allowed users and profiles (e.g. user:profile1|profile2,user2:profile3). Empty allow any identified user`),
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

	if a.identService == nil && a.URL == `` {
		return nil, errors.New(`no authentification target configured`)
	}

	if a.identService != nil {
		retrievedUser, err = a.identService.GetUser(ctx, authContent)
		if err != nil {
			return nil, err
		}
	}

	if a.URL != `` {
		headers := http.Header{}
		headers.Set(authorizationHeader, authContent)

		userBytes, status, _, err := request.Get(ctx, fmt.Sprintf(`%s/user`, a.URL), headers)
		if err != nil {
			if status == http.StatusUnauthorized {
				return nil, ident.ErrEmptyAuth
			}

			return nil, errors.New(`authentication failed: %v`, err)
		}

		retrievedUser = &model.User{}
		if err := json.Unmarshal(userBytes, retrievedUser); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	username := strings.ToLower(retrievedUser.Username)
	if a.users == nil {
		return model.NewUser(retrievedUser.ID, username, retrievedUser.Email, ``), nil
	} else if profiles, ok := a.users[username]; ok {
		return model.NewUser(retrievedUser.ID, username, retrievedUser.Email, profiles), nil
	}

	return nil, ErrForbidden
}

// Handler wrap next authenticated handler
func (a App) Handler(next http.Handler) http.Handler {
	if a.disabled {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		user, err := a.IsAuthenticated(r)
		if err != nil {
			a.onHandlerFail(w, r, err)
			return
		}

		ctx := context.WithValue(r.Context(), ctxUserName, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a App) onHandlerFail(w http.ResponseWriter, r *http.Request, err error) {
	if err == ident.ErrEmptyAuth && a.identService != nil {
		a.identService.OnError(w, r, err)
		return
	}

	if err == ErrForbidden {
		httperror.Forbidden(w)
		return
	}

	httperror.Unauthorized(w, err)
}
