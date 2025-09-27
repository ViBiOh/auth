package github

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/ViBiOh/auth/v2/pkg/ident"
	"github.com/ViBiOh/auth/v2/pkg/middleware"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/flags"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var (
	signMethod       = jwt.SigningMethodHS256
	signValidMethods = []string{signMethod.Alg()}
)

type Service struct {
	config     oauth2.Config
	hmacSecret []byte
}

var _ ident.Provider = Service{}

type Config struct {
	clientID      string
	clientSecret  string
	hmacSecret    string
	redirectURL   string
	jwtExpiration time.Duration
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) *Config {
	var config Config

	flags.New("ClientID", "Client ID").Prefix(prefix).DocPrefix("github").StringVar(fs, &config.clientID, "", overrides)
	flags.New("ClientSecret", "Client Secret").Prefix(prefix).DocPrefix("github").StringVar(fs, &config.clientSecret, "", overrides)
	flags.New("HmacSecret", "HMAC Secret").Prefix(prefix).DocPrefix("github").StringVar(fs, &config.hmacSecret, "", overrides)
	flags.New("JwtExpiration", "JWT Expiration").Prefix(prefix).DocPrefix("github").DurationVar(fs, &config.jwtExpiration, time.Hour*24*5, overrides)
	flags.New("RedirectURL", "URL used for redirection").Prefix(prefix).DocPrefix("github").StringVar(fs, &config.redirectURL, "http://127.0.0.1/auth/github/callback", overrides)

	return &config
}

func New(config *Config) Service {
	return Service{
		config: oauth2.Config{
			ClientID:     config.clientID,
			ClientSecret: config.clientSecret,
			Endpoint:     github.Endpoint,
			RedirectURL:  config.redirectURL,
			Scopes:       nil,
		},
		hmacSecret: []byte(config.hmacSecret),
	}
}

func (s Service) GetUser(ctx context.Context, r *http.Request) (model.User, error) {
	auth, err := r.Cookie("auth")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return model.User{}, middleware.ErrEmptyAuth
		}

		return model.User{}, fmt.Errorf("get auth cookie: %w", err)
	}

	var claim AuthClaims

	if _, err = jwt.ParseWithClaims(auth.Value, &claim, s.jwtKeyFunc, jwt.WithValidMethods(signValidMethods)); err != nil {
		return model.User{}, fmt.Errorf("parse JWT: %w", err)
	}

	return claim.User, nil
}

func (s Service) OnError(http.ResponseWriter, *http.Request, error) {
	panic("unimplemented")
}

func (s Service) jwtKeyFunc(_ *jwt.Token) (any, error) {
	return s.hmacSecret, nil
}
