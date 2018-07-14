package github

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/ViBiOh/auth/pkg/model"
	"github.com/ViBiOh/auth/pkg/provider"
	"github.com/ViBiOh/httputils/pkg/request"
	"github.com/ViBiOh/httputils/pkg/tools"
	"github.com/ViBiOh/httputils/pkg/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type githubEmail struct {
	Email    string
	Primary  bool
	Verified bool
}

type githubUser struct {
	ID    uint
	Login string
}

var (
	userURL  = `https://api.github.com/user`
	emailURL = `https://api.github.com/user/emails`
	endpoint = github.Endpoint
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
	oauthConf *oauth2.Config
	states    sync.Map
}

// NewAuth creates new auth
func NewAuth(config map[string]interface{}) (provider.Auth, error) {
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
		states: sync.Map{},
	}, nil
}

// GetName returns Authorization header prefix
func (*Auth) GetName() string {
	return `GitHub`
}

func (*Auth) getUserEmail(ctx context.Context, header string) string {
	mailResponse, err := request.Get(ctx, emailURL, http.Header{`Authorization`: []string{fmt.Sprintf(`token %s`, header)}})
	if err != nil {
		log.Printf(`[github] Error while fetching email informations: %v: %s`, err, mailResponse)
		return ``
	}

	emails := make([]githubEmail, 0)
	if err := json.Unmarshal(mailResponse, &emails); err != nil {
		log.Printf(`[github] Error while unmarshalling emails informations: %v`, err)
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
func (o *Auth) GetUser(ctx context.Context, header string) (*model.User, error) {
	userResponse, err := request.Get(ctx, userURL, http.Header{`Authorization`: []string{fmt.Sprintf(`token %s`, header)}})
	if err != nil {
		return nil, fmt.Errorf(`Error while fetching user informations: %v`, err)
	}

	user := githubUser{}
	if err := json.Unmarshal(userResponse, &user); err != nil {
		return nil, fmt.Errorf(`Error while unmarshalling user informations: %v`, err)
	}

	return &model.User{ID: user.ID, Username: user.Login, Email: o.getUserEmail(ctx, header)}, nil
}

// Redirect redirects user to GitHub endpoint
func (o *Auth) Redirect() (string, error) {
	state, err := uuid.New()
	o.states.Store(state, true)

	return o.oauthConf.AuthCodeURL(state), err
}

// Login exchanges code for token
func (o *Auth) Login(r *http.Request) (string, error) {
	state := r.FormValue(`state`)
	code := r.FormValue(`code`)

	if _, ok := o.states.Load(state); !ok {
		return ``, provider.ErrInvalidState
	}
	o.states.Delete(state)

	token, err := o.oauthConf.Exchange(oauth2.NoContext, code)
	if err != nil {
		return ``, provider.ErrInvalidCode
	}

	return token.AccessToken, nil
}
