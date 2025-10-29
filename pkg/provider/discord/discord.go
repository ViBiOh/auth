package discord

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ViBiOh/auth/v3/pkg/cookie"
	"github.com/ViBiOh/auth/v3/pkg/model"
	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/httperror"
	"github.com/ViBiOh/httputils/v4/pkg/httpjson"
	"github.com/ViBiOh/httputils/v4/pkg/id"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
)

const (
	verifierCacheKey = "auth:discord:verifier:"
	updateCacheKey   = "auth:discord:update:"
	cookieName       = "_auth"
)

var (
	_ model.Identification = Service{}
	_ model.Authorization  = Service{}
)

type Cache interface {
	Load(ctx context.Context, key string) ([]byte, error)
	Store(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
}

type Provider interface {
	IsAuthorized(ctx context.Context, user model.User, profile string) bool
	GetDiscordUser(ctx context.Context, id, registration string) (model.User, error)
	UpdateDiscordUser(ctx context.Context, user model.User, id, username, avatar string) error
}

type ForbiddenHandler func(http.ResponseWriter, *http.Request, model.User, string)

type Service struct {
	config        oauth2.Config
	cache         Cache
	provider      Provider
	onForbidden   ForbiddenHandler
	onSuccessPath string
	cookie        cookie.Service
}

var _ model.Identification = Service{}

type Config struct {
	clientID      string
	clientSecret  string
	redirectURL   string
	onSuccessPath string
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) *Config {
	var config Config

	flags.New("ClientID", "Client ID").Prefix(prefix).DocPrefix("discord").StringVar(fs, &config.clientID, "", overrides)
	flags.New("ClientSecret", "Client Secret").Prefix(prefix).DocPrefix("discord").StringVar(fs, &config.clientSecret, "", overrides)
	flags.New("RedirectURL", "URL used for redirection").Prefix(prefix).DocPrefix("discord").StringVar(fs, &config.redirectURL, "http://127.0.0.1:1080/auth/discord/callback", overrides)
	flags.New("OnSuccessPath", "Path for redirecting on success").Prefix(prefix).DocPrefix("discord").StringVar(fs, &config.onSuccessPath, "/", overrides)

	return &config
}

func New(config *Config, cache Cache, provider Provider, cookie cookie.Service) Service {
	return Service{
		config: oauth2.Config{
			ClientID:     config.clientID,
			ClientSecret: config.clientSecret,
			Endpoint:     endpoints.Discord,
			RedirectURL:  config.redirectURL,
			Scopes:       []string{"identify"},
		},
		onSuccessPath: config.onSuccessPath,

		cache:    cache,
		provider: provider,
		cookie:   cookie,
	}
}

type Option func(Service) Service

func WithForbiddenHandler(onForbidden ForbiddenHandler) Option {
	return func(instance Service) Service {
		instance.onForbidden = onForbidden

		return instance
	}
}

func (s Service) Register(w http.ResponseWriter, r *http.Request) {
	s.redirect(w, r, r.URL.Query().Get("registration"))
}

func (s Service) redirect(w http.ResponseWriter, r *http.Request, registration string) {
	state := id.New()

	verifier := oauth2.GenerateVerifier()
	payload := State{
		Verifier:     verifier,
		Registration: registration,
	}

	rawPayload, err := json.Marshal(payload)
	if err != nil {
		httperror.InternalServerError(r.Context(), w, fmt.Errorf("marshal state: %w", err))
		return
	}

	if err := s.cache.Store(r.Context(), verifierCacheKey+state, rawPayload, time.Minute*5); err != nil {
		httperror.InternalServerError(r.Context(), w, fmt.Errorf("save state: %w", err))
		return
	}

	http.Redirect(w, r, s.config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier)), http.StatusFound)
}

func (s Service) Callback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	state := verifierCacheKey + r.URL.Query().Get("state")

	rawPayload, err := s.cache.Load(ctx, state)
	if err != nil {
		httperror.NotFound(ctx, w, fmt.Errorf("state not found: %w", err))
		return
	}

	var payload State
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		httperror.NotFound(ctx, w, fmt.Errorf("unmarshal state: %w", err))
		return
	}

	isRegistration := len(payload.Registration) != 0

	oauth2Token, err := s.config.Exchange(ctx, r.URL.Query().Get("code"), oauth2.VerifierOption(payload.Verifier))
	if err != nil {
		httperror.Unauthorized(ctx, w, fmt.Errorf("exchange token: %w", err))
		return
	}

	if err := s.cache.Delete(ctx, state); err != nil {
		httperror.NotFound(ctx, w, fmt.Errorf("delete state: %w", err))
		return
	}

	client := s.config.Client(ctx, oauth2Token)
	resp, err := client.Get("https://discord.com/api/users/@me")
	if err != nil {
		httperror.InternalServerError(ctx, w, fmt.Errorf("get /user: %w", err))
		return
	}

	discordUser, err := httpjson.Read[User](resp)
	if err != nil {
		httperror.InternalServerError(ctx, w, fmt.Errorf("read /user: %w", err))
		return
	}

	user, err := s.provider.GetDiscordUser(ctx, discordUser.ID, payload.Registration)
	if err != nil {
		if errors.Is(err, model.ErrUnknownUser) {
			httperror.NotFound(ctx, w, fmt.Errorf("unregistered user `%s` - `%s`", discordUser.ID, payload.Registration))
			return
		}

		httperror.InternalServerError(ctx, w, fmt.Errorf("get user: %w", err))
		return
	}

	if !s.cookie.Set(ctx, w, oauth2Token, user, cookieName) {
		return
	}

	if isRegistration {
		if err := s.provider.UpdateDiscordUser(ctx, user, discordUser.ID, discordUser.Username, discordUser.Avatar); err != nil {
			httperror.InternalServerError(ctx, w, fmt.Errorf("save github user: %w", err))
			return
		}
	}

	redirectPath := s.onSuccessPath
	if len(payload.Registration) != 0 {
		redirectPath += "?" + url.QueryEscape(payload.Registration)
	}

	w.Header().Add("X-UA-Compatible", "ie=edge")
	w.Header().Add("Content-Type", "text/html; charset=UTF-8")
	w.Header().Add("Cache-Control", "no-cache")

	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, `
<html>
	<head>
		<meta http-equiv="refresh" content=1;url="%[1]s">
	</head>
	<body style="background-color: darkgrey;">
		<img style="display: block; margin: 0 auto; width: 120px;" src="%[2]s">
		<a style="display: block; text-align: center; width: 100vw;" href="%[1]s">Continue...</a>
	</body>
</html>`, redirectPath, discordUser.Image())
}
