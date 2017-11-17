package github

import (
	"encoding/json"
	"flag"
	"fmt"

	"github.com/ViBiOh/auth/auth"
	"github.com/ViBiOh/httputils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

const userURL = `https://api.github.com/user`

type user struct {
	ID    int64
	Login string
}

var (
	state        = flag.String(`githubState`, ``, `GitHub OAuth State`)
	clientID     = flag.String(`githubClientId`, ``, `GitHub OAuth Client ID`)
	clientSecret = flag.String(`githubClientSecret`, ``, `GitHub OAuth Client Secret`)
	oauthConf    *oauth2.Config
)

// Auth auth with GitHub OAuth
type Auth struct{}

// Init provider
func (Auth) Init() error {
	if *clientID != `` {
		oauthConf = &oauth2.Config{
			ClientID:     *clientID,
			ClientSecret: *clientSecret,
			Endpoint:     github.Endpoint,
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
	userResponse, err := httputils.GetBody(userURL, map[string]string{`Authorization`: `token ` + header}, false)
	if err != nil {
		return nil, fmt.Errorf(`Error while fetching user informations: %v`, err)
	}

	user := user{}
	if err := json.Unmarshal(userResponse, &user); err != nil {
		return nil, fmt.Errorf(`Error while unmarshalling user informations: %v`, err)
	}

	return &auth.User{ID: user.ID, Username: user.Login}, nil
}

// GetAccessToken exchange state to token
func (Auth) GetAccessToken(requestState string, requestCode string) (string, error) {
	if *state != requestState {
		return ``, fmt.Errorf(`Invalid state provided for oauth`)
	}

	token, err := oauthConf.Exchange(oauth2.NoContext, requestCode)
	if err != nil {
		return ``, fmt.Errorf(`Invalid code provided for oauth`)
	}

	return token.AccessToken, nil
}
