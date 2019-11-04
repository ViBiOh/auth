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

	"github.com/ViBiOh/auth/pkg/cache"
	"github.com/ViBiOh/auth/pkg/ident"
	"github.com/ViBiOh/auth/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/flags"
	"github.com/ViBiOh/httputils/v3/pkg/httperror"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/httputils/v3/pkg/request"
	"github.com/ViBiOh/httputils/v3/pkg/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var (
	userURL  = "https://api.github.com/user"
	emailURL = "https://api.github.com/user/emails"
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
		clientID:     flags.New(prefix, "github").Name("ClientId").Default("").Label("OAuth Client ID").ToString(fs),
		clientSecret: flags.New(prefix, "github").Name("ClientSecret").Default("").Label("OAuth Client Secret").ToString(fs),
		scopes:       flags.New(prefix, "github").Name("Scopes").Default("").Label("OAuth Scopes, comma separated").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config) (ident.Auth, error) {
	clientID := strings.TrimSpace(*config.clientID)
	if clientID == "" {
		return nil, nil
	}

	var scopes []string
	rawScopes := strings.TrimSpace(*config.scopes)
	if rawScopes != "" {
		scopes = strings.Split(rawScopes, ",")
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
	return "GitHub"
}

func (a *App) hasEmailAccess() bool {
	if len(a.oauthConf.Scopes) == 0 {
		return false
	}

	for _, scope := range a.oauthConf.Scopes {
		if scope == "user:email" || scope == "user" {
			return true
		}
	}

	return false
}

func (a *App) getUserEmail(ctx context.Context, header string) string {
	if !a.hasEmailAccess() {
		return ""
	}

	resp, err := request.New().Get(emailURL).Header("Authorization", fmt.Sprintf("token %s", header)).Send(ctx, nil)
	if err != nil {
		logger.Error("%s", err)
		return ""
	}

	mailResponse, err := request.ReadBodyResponse(resp)
	if err != nil {
		logger.Error("%s", err)
		return ""
	}

	emails := make([]githubEmail, 0)
	if err := json.Unmarshal(mailResponse, &emails); err != nil {
		logger.Error("%s", err)
		return ""
	}

	for _, email := range emails {
		if email.Verified && email.Primary {
			return email.Email
		}
	}

	return ""
}

// GetUser returns User associated to header
func (a *App) GetUser(ctx context.Context, header string) (*model.User, error) {
	if user, ok := a.usersCache.Load(header); ok {
		return user.(*model.User), nil
	}

	resp, err := request.New().Get(userURL).Header("Authorization", fmt.Sprintf("token %s", header)).Send(ctx, nil)
	if err != nil {
		return nil, err
	}

	userResponse, err := request.ReadBodyResponse(resp)
	if err != nil {
		return nil, err
	}

	user := githubUser{}
	if err := json.Unmarshal(userResponse, &user); err != nil {
		return nil, err
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
	state := r.FormValue("state")
	code := r.FormValue("code")

	if _, ok := a.states.Load(state); !ok {
		return "", ident.ErrInvalidState
	}
	a.states.Delete(state)

	token, err := a.oauthConf.Exchange(oauth2.NoContext, code)
	if err != nil {
		return "", ident.ErrInvalidCode
	}

	return token.AccessToken, nil
}

// OnLoginError handle action when login fails
func (*App) OnLoginError(w http.ResponseWriter, r *http.Request, _ error) {
	http.Redirect(w, r, "/redirect/github", http.StatusFound)
}
