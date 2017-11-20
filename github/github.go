package github

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"

	"github.com/ViBiOh/auth/auth"
	"github.com/ViBiOh/httputils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type githubUser struct {
	ID    uint
	Login string
}

var errInvalidState = errors.New(`Invalid state provided for oauth`)
var errCodeState = errors.New(`Invalid code provided for oauth`)

var (
	userURL   = `https://api.github.com/user`
	endpoint  = github.Endpoint
	oauthConf *oauth2.Config
)

var (
	state        = flag.String(`githubState`, ``, `[GitHub] OAuth State`)
	clientID     = flag.String(`githubClientId`, ``, `[GitHub] OAuth Client ID`)
	clientSecret = flag.String(`githubClientSecret`, ``, `[GitHub] OAuth Client Secret`)
)

// Auth auth with GitHub OAuth
type Auth struct{}

// Init provider
func (Auth) Init() error {
	if *clientID != `` {
		oauthConf = &oauth2.Config{
			ClientID:     *clientID,
			ClientSecret: *clientSecret,
			Endpoint:     endpoint,
		}
	}

	return nil
}

// GetName returns Authorization header prefix
func (Auth) GetName() string {
	return `GitHub`
}

// GetUser returns User associated to header
func (Auth) GetUser(header string) (*auth.User, error) {
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

// GetAccessToken exchange code for token
func (Auth) GetAccessToken(requestState string, requestCode string) (string, error) {
	if *state != requestState {
		return ``, errInvalidState
	}

	token, err := oauthConf.Exchange(oauth2.NoContext, requestCode)
	if err != nil {
		return ``, errCodeState
	}

	return token.AccessToken, nil
}
