package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/ViBiOh/auth/v3/pkg/cookie"
	"github.com/ViBiOh/auth/v3/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/httpjson"
	"github.com/ViBiOh/httputils/v4/pkg/id"
	httpModel "github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"golang.org/x/oauth2"
)

const (
	verifierCacheKey = "auth:%s:verifier:"
	updateCacheKey   = "auth:%s:update:"
	cookieName       = "_auth"
)

var _ model.Authentication = Service[ProviderUser[string], string]{}

type Cache interface {
	Load(ctx context.Context, key string) ([]byte, error)
	Store(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
}

type Storage interface {
	DoAtomic(ctx context.Context, action func(context.Context) error) error

	GetInviteByToken(ctx context.Context, token string) (model.User, error)
	Delete(ctx context.Context, user model.User) error
	DeleteInvite(ctx context.Context, user model.User) error
}

type ProviderUser[I comparable] interface {
	GetID() I
}

type (
	LinkHandler                                    func(ctx context.Context, old, new model.User) error
	CreateHandler[T ProviderUser[I], I comparable] func(ctx context.Context, invite model.User, user T) (model.User, error)
	GetHandler[I comparable]                       func(ctx context.Context, id I) (model.User, error)
)

type Service[T ProviderUser[I], I comparable] struct {
	config        oauth2.Config
	cache         Cache
	storage       Storage
	renderer      *renderer.Service
	linkHandler   LinkHandler
	createHandler CreateHandler[T, I]
	getHandler    GetHandler[I]
	name          string
	getURL        string
	onSuccessPath string
	cookie        cookie.Service[model.OAuthClaim]
}

var _ model.Authentication = Service[ProviderUser[string], string]{}

func New[T ProviderUser[I], I comparable](name, getURL, onSuccessPath string, config oauth2.Config, cache Cache, storage Storage, linkHandler LinkHandler, createHandler CreateHandler[T, I], getHandler GetHandler[I], renderer *renderer.Service, cookie cookie.Service[model.OAuthClaim]) Service[T, I] {
	return Service[T, I]{
		name:          name,
		getURL:        getURL,
		onSuccessPath: onSuccessPath,
		config:        config,

		cache:         cache,
		storage:       storage,
		linkHandler:   linkHandler,
		createHandler: createHandler,
		getHandler:    getHandler,
		renderer:      renderer,
		cookie:        cookie,
	}
}

func (s Service[T, I]) Mux(prefix string, mux *http.ServeMux) {
	mux.HandleFunc(prefix+"/logout", s.Logout)
	mux.HandleFunc(prefix+"/register", s.Register)
	mux.HandleFunc(prefix+"/callback", s.Callback)
}

func (s Service[T, I]) Logout(w http.ResponseWriter, r *http.Request) {
	s.cookie.Clear(w, cookieName)

	s.renderer.Serve(w, r, renderer.NewPage("auth", http.StatusOK, map[string]any{
		"Redirect": "/",
		"Message":  renderer.NewSuccessMessage("Logout success!"),
	}))
}

func (s Service[T, I]) Register(w http.ResponseWriter, r *http.Request) {
	s.redirect(w, r, r.URL.Query().Get("registration"), r.URL.Query().Get("redirect"))
}

func (s Service[T, I]) redirect(w http.ResponseWriter, r *http.Request, registration, redirect string) {
	ctx := r.Context()
	state := id.New()

	if len(registration) != 0 {
		if _, err := s.storage.GetInviteByToken(ctx, registration); err != nil && errors.Is(err, model.ErrUnknownUser) {
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

func (s Service[T, I]) Callback(w http.ResponseWriter, r *http.Request) {
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
	resp, err := client.Get(s.getURL)
	if err != nil {
		s.renderer.Error(w, r, nil, fmt.Errorf("get user from provider: %w", err))
		return
	}

	providerUser, err := httpjson.Read[T](resp)
	if err != nil {
		s.renderer.Error(w, r, nil, fmt.Errorf("read user from provider: %w", err))
		return
	}

	redirect := payload.Redirection
	if len(redirect) == 0 {
		redirect = s.onSuccessPath
	}

	isRegistration := len(payload.Registration) != 0

	user, err := s.getHandler(ctx, providerUser.GetID())
	if err == nil && !isRegistration {
		s.callbackSuccess(ctx, w, r, state, oauth2Token, user, redirect)
		return
	}

	if err != nil && !errors.Is(err, model.ErrUnknownUser) {
		s.renderer.Error(w, r, nil, fmt.Errorf("get user: %w", err))
		return
	}

	invite, err := s.storage.GetInviteByToken(ctx, payload.Registration)
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

	if err := s.storage.DoAtomic(ctx, func(ctx context.Context) (err error) {
		if len(user.ID) == 0 {
			user, err = s.createHandler(ctx, invite, providerUser)
			if err != nil {
				return err
			}
		}

		if err := s.linkHandler(ctx, invite, user); err != nil {
			return fmt.Errorf("invite handler: %w", err)
		}

		if user.ID != invite.ID {
			return s.storage.Delete(ctx, invite)
		}

		return s.storage.DeleteInvite(ctx, invite)
	}); err != nil {
		s.renderer.Error(w, r, nil, fmt.Errorf("upsert user: %w", err))
		return
	}

	s.callbackSuccess(ctx, w, r, state, oauth2Token, user, redirect)
}

func (s Service[T, I]) callbackSuccess(ctx context.Context, w http.ResponseWriter, r *http.Request, state string, oauth2Token *oauth2.Token, user model.User, redirect string) {
	if err := s.cache.Delete(ctx, state); err != nil {
		slog.ErrorContext(ctx, "unable to delete state", slog.Any("error", err))
	}

	if !s.cookie.Set(ctx, w, cookieName, model.OAuthClaim{Token: oauth2Token, User: user}) {
		return
	}

	s.renderer.Serve(w, r, renderer.NewPage("auth", http.StatusOK, map[string]any{
		"Redirect": redirect,
		"Image":    user.Image,
		"Message":  renderer.NewSuccessMessage("Login success!"),
	}))
}
