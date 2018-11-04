package github

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
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

// Flags add flags for given prefix
func Flags(prefix string) map[string]interface{} {
	return map[string]interface{}{
		`clientID`:     flag.String(tools.ToCamel(fmt.Sprintf(`%sClientId`, prefix)), ``, `[GitHub] OAuth Client ID`),
		`clientSecret`: flag.String(tools.ToCamel(fmt.Sprintf(`%sClientSecret`, prefix)), ``, `[GitHub] OAuth Client Secret`),
		`scopes`:       flag.String(tools.ToCamel(fmt.Sprintf(`%sScopes`, prefix)), ``, `[GitHub] OAuth Scopes, comma separated`),
	}
}

// Auth auth with GitHub OAuth
type Auth struct {
	oauthConf  *oauth2.Config
	usersCache cache.TimeMap
	states     sync.Map
}

// NewAuth creates new auth
func NewAuth(config map[string]interface{}) (ident.Auth, error) {
	clientID := strings.TrimSpace(*(config[`clientID`].(*string)))
	if clientID == `` {
		return nil, nil
	}

	var scopes []string
	rawScopes := strings.TrimSpace(*(config[`scopes`].(*string)))
	if rawScopes != `` {
		scopes = strings.Split(rawScopes, `,`)
	}

	return &Auth{
		oauthConf: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: strings.TrimSpace(*(config[`clientSecret`].(*string))),
			Endpoint:     endpoint,
			Scopes:       scopes,
		},
		usersCache: cache.New(),
		states:     sync.Map{},
	}, nil
}

// GetName returns Authorization header prefix
func (*Auth) GetName() string {
	return `GitHub`
}

func (a *Auth) hasEmailAccess() bool {
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

func (a *Auth) getUserEmail(ctx context.Context, header string) string {
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
func (a *Auth) GetUser(ctx context.Context, header string) (*model.User, error) {
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

	githubUser := &model.User{ID: user.ID, Username: user.Login, Email: a.getUserEmail(ctx, header)}
	a.usersCache.Store(header, githubUser, userCacheDuration)

	return githubUser, nil
}

// Redirect redirects user to GitHub endpoint
func (a *Auth) Redirect(w http.ResponseWriter, r *http.Request) {
	state, err := uuid.New()
	if err != nil {
		httperror.InternalServerError(w, err)
	}

	a.states.Store(state, true)
	http.Redirect(w, r, a.oauthConf.AuthCodeURL(state), http.StatusFound)
}

// Login exchanges code for token
func (a *Auth) Login(r *http.Request) (string, error) {
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
func (*Auth) OnLoginError(w http.ResponseWriter, r *http.Request, _ error) {
	http.Redirect(w, r, `/redirect/github`, http.StatusFound)
}
