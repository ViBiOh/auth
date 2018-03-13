package twitter

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ViBiOh/auth/provider"
	"github.com/ViBiOh/httputils/request"
	"github.com/ViBiOh/httputils/tools"
	"github.com/ViBiOh/httputils/uuid"
	"golang.org/x/oauth2"
)

type twitterUser struct {
	ID    uint
	Login string
}

var (
	userURL  = `https://api.twitter.com/1.1/account/settings`
	endpoint = oauth2.Endpoint{
		AuthURL:  `https://api.twitter.com/oauth/authorize`,
		TokenURL: `https://api.twitter.com/oauth/access_token`,
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
	states    sync.Map
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
			states: sync.Map{},
		}, nil
	}

	return nil, nil
}

// GetName returns Authorization header prefix
func (*Auth) GetName() string {
	return `Twitter`
}

// GetUser returns User associated to header
func (a *Auth) GetUser(header string) (*provider.User, error) {
	nowTS := time.Now().Unix()

	userResponse, err := request.Get(userURL, map[string]string{`Authorization`: fmt.Sprintf(`
OAuth oauth_consumer_key="%s",
oauth_nonce="%s",
oauth_signature="%s",
oauth_signature_method="HMAC-SHA1",
oauth_timestamp="%d",
oauth_token="%s",
oauth_version="1.0"`, a.oauthConf.ClientID, tools.Sha1(nowTS), ``, nowTS, header)})

	if err != nil {
		return nil, fmt.Errorf(`Error while fetching user informations: %v`, err)
	}

	user := twitterUser{}
	if err := json.Unmarshal(userResponse, &user); err != nil {
		return nil, fmt.Errorf(`Error while unmarshalling user informations: %v`, err)
	}

	return &provider.User{ID: user.ID, Username: user.Login}, nil
}

// Redirect redirects user to Twitter endpoint
func (a *Auth) Redirect() (string, error) {
	state, err := uuid.New()
	a.states.Store(state, true)

	return a.oauthConf.AuthCodeURL(state), err
}

// Login exchanges code for token
func (a *Auth) Login(r *http.Request) (string, error) {
	state := r.FormValue(`state`)
	code := r.FormValue(`code`)

	if _, ok := a.states.Load(state); !ok {
		return ``, provider.ErrInvalidState
	}
	a.states.Delete(state)

	token, err := a.oauthConf.Exchange(oauth2.NoContext, code)
	if err != nil {
		return ``, provider.ErrInvalidCode
	}

	return token.AccessToken, nil
}
