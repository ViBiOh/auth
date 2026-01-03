package github

import (
	"context"
	"flag"

	"github.com/ViBiOh/auth/v3/pkg/cookie"
	"github.com/ViBiOh/auth/v3/pkg/model"
	"github.com/ViBiOh/auth/v3/pkg/provider/oauth"
	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type Config struct {
	clientID      string
	clientSecret  string
	redirectURL   string
	onSuccessPath string
}

type Storage interface {
	oauth.Storage

	CreateGithub(context.Context, model.User, model.GitHubUser) (model.User, error)
	GetGitHubUser(context.Context, uint64) (model.User, error)
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) *Config {
	var config Config

	flags.New("ClientID", "Client ID").Prefix(prefix).DocPrefix("github").StringVar(fs, &config.clientID, "", overrides)
	flags.New("ClientSecret", "Client Secret").Prefix(prefix).DocPrefix("github").StringVar(fs, &config.clientSecret, "", overrides)
	flags.New("RedirectURL", "URL used for redirection").Prefix(prefix).DocPrefix("github").StringVar(fs, &config.redirectURL, "http://127.0.0.1:1080/oauth/github/callback", overrides)
	flags.New("OnSuccessPath", "Path for redirecting on success").Prefix(prefix).DocPrefix("github").StringVar(fs, &config.onSuccessPath, "/", overrides)

	return &config
}

func New(config *Config, cache oauth.Cache, storage Storage, linkHandler oauth.LinkHandler, renderer *renderer.Service, cookie cookie.Service) oauth.Service[model.GitHubUser, uint64] {
	return oauth.New("github", "https://api.github.com/user", config.onSuccessPath, oauth2.Config{
		ClientID:     config.clientID,
		ClientSecret: config.clientSecret,
		Endpoint:     github.Endpoint,
		RedirectURL:  config.redirectURL,
	}, cache, storage, linkHandler, storage.CreateGithub, storage.GetGitHubUser, renderer, cookie)
}
