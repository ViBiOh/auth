package google

import (
	"context"
	"flag"

	"github.com/ViBiOh/auth/v3/pkg/cookie"
	"github.com/ViBiOh/auth/v3/pkg/model"
	"github.com/ViBiOh/auth/v3/pkg/provider/oauth"
	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Config struct {
	clientID      string
	clientSecret  string
	redirectURL   string
	onSuccessPath string
}

type Storage interface {
	oauth.Storage

	CreateGoogle(context.Context, model.User, model.GoogleUser) (model.User, error)
	GetGoogleUser(context.Context, string) (model.User, error)
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) *Config {
	var config Config

	flags.New("ClientID", "Client ID").Prefix(prefix).DocPrefix("google").StringVar(fs, &config.clientID, "", overrides)
	flags.New("ClientSecret", "Client Secret").Prefix(prefix).DocPrefix("google").StringVar(fs, &config.clientSecret, "", overrides)
	flags.New("RedirectURL", "URL used for redirection").Prefix(prefix).DocPrefix("google").StringVar(fs, &config.redirectURL, "http://127.0.0.1:1080/oauth/google/callback", overrides)
	flags.New("OnSuccessPath", "Path for redirecting on success").Prefix(prefix).DocPrefix("google").StringVar(fs, &config.onSuccessPath, "/", overrides)

	return &config
}

func New(config *Config, cache oauth.Cache, storage Storage, linkHandler oauth.LinkHandler, renderer *renderer.Service, cookie cookie.Service[model.OAuthClaim]) oauth.Service[model.GoogleUser, string] {
	return oauth.New("google", "https://www.googleapis.com/oauth2/v3/userinfo", config.onSuccessPath, oauth2.Config{
		ClientID:     config.clientID,
		ClientSecret: config.clientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  config.redirectURL,
		Scopes:       []string{"openid", "profile"},
	}, cache, storage, linkHandler, storage.CreateGoogle, storage.GetGoogleUser, renderer, cookie)
}
