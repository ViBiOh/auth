package twitter

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/ViBiOh/auth/pkg/model"
	"github.com/ViBiOh/auth/pkg/provider"
	"github.com/ViBiOh/httputils/pkg/request"
	"github.com/ViBiOh/httputils/pkg/tools"
	"golang.org/x/oauth2"
)

var (
	endpoint = oauth2.Endpoint{
		AuthURL:  `https://api.twitter.com/oauth2/authorize`,
		TokenURL: `https://api.twitter.com/oauth2/token`,
	}
)

// Flags add flags for given prefix
func Flags(prefix string) map[string]interface{} {
	return map[string]interface{}{
		`key`:    flag.String(tools.ToCamel(fmt.Sprintf(`%s%s`, prefix, `Key`)), ``, `[Twitter] Consumer Key`),
		`secret`: flag.String(tools.ToCamel(fmt.Sprintf(`%s%s`, prefix, `Secret`)), ``, `[Twitter] Consumer Secret`),
	}
}

// Auth auth with Twitter OAuth
type Auth struct {
	oauthConf *oauth2.Config
}

// NewAuth creates new auth
func NewAuth(config map[string]interface{}) (provider.Auth, error) {
	if key, ok := config[`key`]; ok && *(key.(*string)) != `` {
		log.Print(`Twitter provider implementation is WIP`)

		return &Auth{
			oauthConf: &oauth2.Config{
				ClientID:     *(key.(*string)),
				ClientSecret: *(config[`secret`].(*string)),
				Endpoint:     endpoint,
			},
		}, nil
	}

	return nil, nil
}

// GetName returns Authorization header prefix
func (*Auth) GetName() string {
	return `Twitter`
}

// GetUser returns User associated to header
func (a *Auth) GetUser(header string) (*model.User, error) {
	return nil, errors.New(`WIP`)
}

// Redirect redirects user to Twitter endpoint
func (a *Auth) Redirect() (string, error) {
	return a.oauthConf.AuthCodeURL(``), nil
}

// Login exchanges code for token
func (a *Auth) Login(r *http.Request) (string, error) {
	token, err := a.oauthConf.Exchange(oauth2.NoContext, request.GetBasicAuth(a.oauthConf.ClientID, a.oauthConf.ClientSecret))
	if err != nil {
		return ``, provider.ErrInvalidCode
	}

	return token.AccessToken, nil
}
