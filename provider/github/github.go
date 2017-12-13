package github

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"sync"

	"github.com/ViBiOh/auth/auth"
	"github.com/ViBiOh/auth/provider"
	"github.com/ViBiOh/httputils"
	"github.com/ViBiOh/httputils/tools"
	"github.com/ViBiOh/httputils/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type githubUser struct {
	ID    uint
	Login string
}

var (
	userURL  = `https://api.github.com/user`
	endpoint = github.Endpoint
)

// Flags add flags for given prefix
func Flags(prefix string) map[string]interface{} {
	return map[string]interface{}{
		`clientID`:     flag.String(tools.ToCamel(prefix+`ClientId`), ``, `[GitHub] OAuth Client ID`),
		`clientSecret`: flag.String(tools.ToCamel(prefix+`ClientSecret`), ``, `[GitHub] OAuth Client Secret`),
	}
}

// Auth auth with GitHub OAuth
type Auth struct {
	oauthConf *oauth2.Config
	states    sync.Map
}

// Init provider
func (o *Auth) Init(config map[string]interface{}) error {
	if clientID, ok := config[`clientID`]; ok && *(clientID.(*string)) != `` {
		o.oauthConf = &oauth2.Config{
			ClientID:     *(clientID.(*string)),
			ClientSecret: *(config[`clientSecret`].(*string)),
			Endpoint:     endpoint,
		}
		o.states = sync.Map{}
	}

	return nil
}

// GetName returns Authorization header prefix
func (*Auth) GetName() string {
	return `GitHub`
}

// GetUser returns User associated to header
func (*Auth) GetUser(header string) (*auth.User, error) {
	userResponse, err := httputils.GetBody(userURL, map[string]string{`Authorization`: `token ` + header})
	if err != nil {
		return nil, fmt.Errorf(`Error while fetching user informations: %v`, err)
	}

	user := githubUser{}
	if err := json.Unmarshal(userResponse, &user); err != nil {
		return nil, fmt.Errorf(`Error while unmarshalling user informations: %v`, err)
	}

	return &auth.User{ID: user.ID, Username: user.Login}, nil
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
