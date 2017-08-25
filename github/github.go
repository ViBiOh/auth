package github

import (
	"encoding/json"
	"flag"
	"fmt"

	"github.com/ViBiOh/httputils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

const userURL = `https://api.github.com/user`

type user struct {
	Login string `json:"login"`
}

var (
	state        = flag.String(`githubState`, ``, `GitHub OAuth State`)
	clientID     = flag.String(`githubClientId`, ``, `GitHub OAuth Client ID`)
	clientSecret = flag.String(`githubClientSecret`, ``, `GitHub OAuth Client Secret`)
	oauthConf    *oauth2.Config
)

// Init configuration
func Init() error {
	if *clientID != `` {
		oauthConf = &oauth2.Config{
			ClientID:     *clientID,
			ClientSecret: *clientSecret,
			Endpoint:     github.Endpoint,
		}
	}

	return nil
}

const maxConcurrentAuth = 32

var tokenPool = make(chan int, maxConcurrentAuth)

func getToken() {
	tokenPool <- 1
}

func releaseToken() {
	<-tokenPool
}

// GetAccessToken returns access token for given state and code
func GetAccessToken(requestState string, requestCode string) (string, error) {
	getToken()
	defer releaseToken()

	if *state != requestState {
		return ``, fmt.Errorf(`Invalid state provided for oauth`)
	}

	token, err := oauthConf.Exchange(oauth2.NoContext, requestCode)
	if err != nil {
		return ``, fmt.Errorf(`Invalid code provided for oauth`)
	}

	return token.AccessToken, nil
}

// GetUsername returns username of given token
func GetUsername(token string) (string, error) {
	getToken()
	defer releaseToken()

	userResponse, err := httputils.GetBody(userURL, `token `+token, false)
	if err != nil {
		return ``, fmt.Errorf(`Error while fetching user informations: %v`, err)
	}

	user := user{}
	if err := json.Unmarshal(userResponse, &user); err != nil {
		return ``, fmt.Errorf(`Error while unmarshalling user informations: %v`, err)
	}

	return user.Login, nil
}
