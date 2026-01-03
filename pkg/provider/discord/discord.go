package discord

import (
	"context"
	"flag"

	"github.com/ViBiOh/auth/v3/pkg/cookie"
	"github.com/ViBiOh/auth/v3/pkg/model"
	"github.com/ViBiOh/auth/v3/pkg/provider/oauth"
	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
)

type Config struct {
	clientID      string
	clientSecret  string
	redirectURL   string
	onSuccessPath string
}

type Storage interface {
	oauth.Storage

	CreateDiscord(context.Context, model.User, model.DiscordUser) (model.User, error)
	GetDiscordUser(context.Context, string) (model.User, error)
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) *Config {
	var config Config

	flags.New("ClientID", "Client ID").Prefix(prefix).DocPrefix("discord").StringVar(fs, &config.clientID, "", overrides)
	flags.New("ClientSecret", "Client Secret").Prefix(prefix).DocPrefix("discord").StringVar(fs, &config.clientSecret, "", overrides)
	flags.New("RedirectURL", "URL used for redirection").Prefix(prefix).DocPrefix("discord").StringVar(fs, &config.redirectURL, "http://127.0.0.1:1080/oauth/discord/callback", overrides)
	flags.New("OnSuccessPath", "Path for redirecting on success").Prefix(prefix).DocPrefix("discord").StringVar(fs, &config.onSuccessPath, "/", overrides)

	return &config
}

func New(config *Config, cache oauth.Cache, storage Storage, linkHandler oauth.LinkHandler, renderer *renderer.Service, cookie cookie.Service) oauth.Service[model.DiscordUser, string] {
	return oauth.New("discord", "https://discord.com/api/users/@me", config.onSuccessPath, oauth2.Config{
		ClientID:     config.clientID,
		ClientSecret: config.clientSecret,
		Endpoint:     endpoints.Discord,
		RedirectURL:  config.redirectURL,
		Scopes:       []string{"identify"},
	}, cache, storage, linkHandler, storage.CreateDiscord, storage.GetDiscordUser, renderer, cookie)
}
