package discord

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/ViBiOh/auth/v3/pkg/cookie"
	"github.com/ViBiOh/auth/v3/pkg/model"
	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/httpjson"
	"github.com/ViBiOh/httputils/v4/pkg/id"
	httpModel "github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
)

const (
	verifierCacheKey = "auth:discord:verifier:"
	updateCacheKey   = "auth:discord:update:"
	cookieName       = "_auth"
)

var _ model.Authentication = Service{}

type Cache interface {
	Load(ctx context.Context, key string) ([]byte, error)
	Store(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
}

type Provider interface {
	DoAtomic(ctx context.Context, action func(context.Context) error) error

	CreateDiscord(ctx context.Context, invite model.User, id, username, avatar string) (model.User, error)
	GetDiscordUser(ctx context.Context, id string) (model.User, error)

	GetInviteByToken(ctx context.Context, token string) (model.User, error)
	Delete(ctx context.Context, user model.User) error
	DeleteInvite(ctx context.Context, user model.User) error
}

type (
	LinkHandler      func(ctx context.Context, old, new model.User) error
	ForbiddenHandler func(http.ResponseWriter, *http.Request, model.User, string)
)

type Service struct {
	config        oauth2.Config
	cache         Cache
	provider      Provider
	linkHandler   LinkHandler
	onSuccessPath string
	renderer      *renderer.Service
	cookie        cookie.Service
}

var _ model.Authentication = Service{}

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
	flags.New("RedirectURL", "URL used for redirection").Prefix(prefix).DocPrefix("discord").StringVar(fs, &config.redirectURL, "http://127.0.0.1:1080/oauth/discord/callback", overrides)
	flags.New("OnSuccessPath", "Path for redirecting on success").Prefix(prefix).DocPrefix("discord").StringVar(fs, &config.onSuccessPath, "/", overrides)

	return &config
}

func New(config *Config, cache Cache, provider Provider, linkHandler LinkHandler, renderer *renderer.Service, cookie cookie.Service) Service {
	return Service{
		config: oauth2.Config{
			ClientID:     config.clientID,
			ClientSecret: config.clientSecret,
			Endpoint:     endpoints.Discord,
			RedirectURL:  config.redirectURL,
			Scopes:       []string{"identify"},
		},
		onSuccessPath: config.onSuccessPath,

		cache:       cache,
		provider:    provider,
		linkHandler: linkHandler,
		renderer:    renderer,
		cookie:      cookie,
	}
}

func (s Service) Mux(prefix string, mux *http.ServeMux) {
	mux.HandleFunc(prefix+"/logout", s.Logout)
	mux.HandleFunc(prefix+"/register", s.Register)
	mux.HandleFunc(prefix+"/callback", s.Callback)
}

func (s Service) Logout(w http.ResponseWriter, r *http.Request) {
	s.cookie.Clear(w, cookieName)

	s.renderer.Serve(w, r, renderer.NewPage("auth", http.StatusOK, map[string]any{
		"Redirect": "/",
		"Message":  renderer.NewSuccessMessage("Logout success!"),
	}))
}

func (s Service) Register(w http.ResponseWriter, r *http.Request) {
	s.redirect(w, r, r.URL.Query().Get("registration"), r.URL.Query().Get("redirect"))
}

func (s Service) redirect(w http.ResponseWriter, r *http.Request, registration, redirect string) {
	ctx := r.Context()
	state := id.New()

	if len(registration) != 0 {
		if _, err := s.provider.GetInviteByToken(ctx, registration); err != nil && errors.Is(err, model.ErrUnknownUser) {
			s.renderer.Serve(w, r, renderer.NewPage("auth", http.StatusOK, map[string]any{
				"Redirect": redirect,
				"Message":  renderer.NewErrorMessage("Unknown registration code or already used"),
			}))
			return
		}
	}

	verifier := oauth2.GenerateVerifier()
	payload := State{
		Verifier:     verifier,
		Registration: registration,
		Redirection:  redirect,
	}

	rawPayload, err := json.Marshal(payload)
	if err != nil {
		s.renderer.Error(w, r, nil, fmt.Errorf("marshal state: %w", err))
		return
	}

	if err := s.cache.Store(ctx, verifierCacheKey+state, rawPayload, time.Minute*5); err != nil {
		s.renderer.Error(w, r, nil, fmt.Errorf("save state: %w", err))
		return
	}

	http.Redirect(w, r, s.config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier)), http.StatusFound)
}

func (s Service) Callback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	state := verifierCacheKey + r.URL.Query().Get("state")

	rawPayload, err := s.cache.Load(ctx, state)
	if err != nil {
		s.renderer.Error(w, r, nil, httpModel.WrapNotFound(fmt.Errorf("state not found: %w", err)))
		return
	}

	var payload State
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		s.renderer.Error(w, r, nil, fmt.Errorf("unmarshal state: %w", err))
		return
	}

	oauth2Token, err := s.config.Exchange(ctx, r.URL.Query().Get("code"), oauth2.VerifierOption(payload.Verifier))
	if err != nil {
		s.renderer.Error(w, r, nil, fmt.Errorf("exchange token: %w", err))
		return
	}

	client := s.config.Client(ctx, oauth2Token)
	resp, err := client.Get("https://discord.com/api/users/@me")
	if err != nil {
		s.renderer.Error(w, r, nil, fmt.Errorf("get discord /user: %w", err))
		return
	}

	discordUser, err := httpjson.Read[User](resp)
	if err != nil {
		s.renderer.Error(w, r, nil, fmt.Errorf("read discord /user: %w", err))
		return
	}

	redirect := payload.Redirection
	if len(redirect) == 0 {
		redirect = s.onSuccessPath
	}

	isRegistration := len(payload.Registration) != 0

	user, err := s.provider.GetDiscordUser(ctx, discordUser.ID)
	if err == nil && !isRegistration {
		s.callbackSuccess(ctx, w, r, state, oauth2Token, user, redirect)
		return
	}

	if err != nil && !errors.Is(err, model.ErrUnknownUser) {
		s.renderer.Error(w, r, nil, fmt.Errorf("get user: %w", err))
		return
	}

	invite, err := s.provider.GetInviteByToken(ctx, payload.Registration)
	if err != nil {
		if errors.Is(err, model.ErrUnknownUser) {
			s.renderer.Serve(w, r, renderer.NewPage("auth", http.StatusOK, map[string]any{
				"Redirect": redirect,
				"Message":  renderer.NewErrorMessage("Unknown registration code or already used"),
			}))
			return
		}

		s.renderer.Error(w, r, nil, fmt.Errorf("get registration: %w", err))
		return
	}

	if err := s.provider.DoAtomic(ctx, func(ctx context.Context) (err error) {
		if len(user.ID) == 0 {
			user, err = s.provider.CreateDiscord(ctx, invite, discordUser.ID, discordUser.Username, discordUser.Avatar)
			if err != nil {
				return err
			}
		}

		if err := s.linkHandler(ctx, invite, user); err != nil {
			return fmt.Errorf("invite handler: %w", err)
		}

		if user.ID != invite.ID {
			return s.provider.Delete(ctx, invite)
		}

		return s.provider.DeleteInvite(ctx, invite)
	}); err != nil {
		s.renderer.Error(w, r, nil, fmt.Errorf("upsert user: %w", err))
		return
	}

	s.callbackSuccess(ctx, w, r, state, oauth2Token, user, redirect)
}

func (s Service) callbackSuccess(ctx context.Context, w http.ResponseWriter, r *http.Request, state string, oauth2Token *oauth2.Token, user model.User, redirect string) {
	if err := s.cache.Delete(ctx, state); err != nil {
		slog.ErrorContext(ctx, "unable to delete state", slog.Any("error", err))
	}

	if !s.cookie.Set(ctx, w, oauth2Token, user, cookieName) {
		return
	}

	s.renderer.Serve(w, r, renderer.NewPage("auth", http.StatusOK, map[string]any{
		"Redirect": redirect,
		"Image":    user.Image,
		"Message":  renderer.NewSuccessMessage("Login success!"),
	}))
}
