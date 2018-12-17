package github

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ViBiOh/auth/pkg/ident"
	"github.com/ViBiOh/auth/pkg/model"
	"github.com/ViBiOh/httputils/pkg/cache"
	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/httputils/pkg/httperror"
	"github.com/ViBiOh/httputils/pkg/logger"
	"github.com/ViBiOh/httputils/pkg/request"
	"github.com/ViBiOh/httputils/pkg/tools"
	"github.com/ViBiOh/httputils/pkg/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var (
	userURL  = `https://api.github.com/user`
	emailURL = `https://api.github.com/user/emails`
	endpoint = github.Endpoint

	userCacheDuration = time.Minute * 5
)

// Config of package
type Config struct {
	clientID     *string
	clientSecret *string
	scopes       *string
}

// App of package
type App struct {
	oauthConf  *oauth2.Config
	states     sync.Map
	usersCache cache.TimeMap
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		clientID:     fs.String(tools.ToCamel(fmt.Sprintf(`%sClientId`, prefix)), ``, `[github] OAuth Client ID`),
		clientSecret: fs.String(tools.ToCamel(fmt.Sprintf(`%sClientSecret`, prefix)), ``, `[github] OAuth Client Secret`),
		scopes:       fs.String(tools.ToCamel(fmt.Sprintf(`%sScopes`, prefix)), ``, `[github] OAuth Scopes, comma separated`),
	}
}

// New creates new App from Config
func New(config Config) (ident.Auth, error) {
	clientID := strings.TrimSpace(*config.clientID)
	if clientID == `` {
		return nil, nil
	}

	var scopes []string
	rawScopes := strings.TrimSpace(*config.scopes)
	if rawScopes != `` {
		scopes = strings.Split(rawScopes, `,`)
	}

	return &App{
		oauthConf: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: strings.TrimSpace(*config.clientSecret),
			Endpoint:     endpoint,
			Scopes:       scopes,
		},
		usersCache: cache.New(),
		states:     sync.Map{},
	}, nil
}

// GetName returns Authorization header prefix
func (*App) GetName() string {
	return `GitHub`
}

func (a *App) hasEmailAccess() bool {
	if len(a.oauthConf.Scopes) == 0 {
		return false
	}

	for _, scope := range a.oauthConf.Scopes {
		if scope == `user:email` || scope == `user` {
			return true
		}
	}

	return false
}

func (a *App) getUserEmail(ctx context.Context, header string) string {
	if !a.hasEmailAccess() {
		return ``
	}

	mailResponse, _, _, err := request.Get(ctx, emailURL, http.Header{`Authorization`: []string{fmt.Sprintf(`token %s`, header)}})
	if err != nil {
		logger.Error(`%+v`, err)
		return ``
	}

	emails := make([]githubEmail, 0)
	if err := json.Unmarshal(mailResponse, &emails); err != nil {
		logger.Error(`%+v`, errors.WithStack(err))
		return ``
	}

	for _, email := range emails {
		if email.Verified && email.Primary {
			return email.Email
		}
	}

	return ``
}

// GetUser returns User associated to header
func (a *App) GetUser(ctx context.Context, header string) (*model.User, error) {
	if user, ok := a.usersCache.Load(header); ok {
		return user.(*model.User), nil
	}

	userResponse, _, _, err := request.Get(ctx, userURL, http.Header{`Authorization`: []string{fmt.Sprintf(`token %s`, header)}})
	if err != nil {
		return nil, err
	}

	user := githubUser{}
	if err := json.Unmarshal(userResponse, &user); err != nil {
		return nil, errors.WithStack(err)
	}

	githubUser := &model.User{ID: strconv.Itoa(user.ID), Username: user.Login, Email: a.getUserEmail(ctx, header)}
	a.usersCache.Store(header, githubUser, userCacheDuration)

	return githubUser, nil
}

// Redirect redirects user to GitHub endpoint
func (a *App) Redirect(w http.ResponseWriter, r *http.Request) {
	state, err := uuid.New()
	if err != nil {
		httperror.InternalServerError(w, err)
		return
	}

	a.states.Store(state, true)
	http.Redirect(w, r, a.oauthConf.AuthCodeURL(state), http.StatusFound)
}

// Login exchanges code for token
func (a *App) Login(r *http.Request) (string, error) {
	state := r.FormValue(`state`)
	code := r.FormValue(`code`)

	if _, ok := a.states.Load(state); !ok {
		return ``, ident.ErrInvalidState
	}
	a.states.Delete(state)

	token, err := a.oauthConf.Exchange(oauth2.NoContext, code)
	if err != nil {
		return ``, ident.ErrInvalidCode
	}

	return token.AccessToken, nil
}

// OnLoginError handle action when login fails
func (*App) OnLoginError(w http.ResponseWriter, r *http.Request, _ error) {
	http.Redirect(w, r, `/redirect/github`, http.StatusFound)
}
