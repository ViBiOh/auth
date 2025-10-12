package github

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/httperror"
	"github.com/ViBiOh/httputils/v4/pkg/httpjson"
	"github.com/ViBiOh/httputils/v4/pkg/id"
	httpmodel "github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

const (
	verifierCacheKey = "auth:github:verifier:"
	cookieName       = "_auth"
)

var (
	_ model.Identification = Service{}
	_ model.Authorization  = Service{}
)

var (
	signMethod       = jwt.SigningMethodHS256
	signValidMethods = []string{signMethod.Alg()}
)

type Cache interface {
	Load(ctx context.Context, key string) ([]byte, error)
	Store(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
}

type Provider interface {
	IsAuthorized(ctx context.Context, user model.User, profile string) bool
	GetGitHubUser(ctx context.Context, registration string) (model.User, error)
	UpdateGitHubUser(ctx context.Context, user model.User, githubID, githubLogin string) error
}

type ForbiddenHandler func(http.ResponseWriter, *http.Request, model.User, string)

type Service struct {
	config        oauth2.Config
	cache         Cache
	provider      Provider
	onForbidden   ForbiddenHandler
	onSuccessPath string
	hmacSecret    []byte
	jwtExpiration time.Duration
}

var _ model.Identification = Service{}

type Config struct {
	clientID      string
	clientSecret  string
	hmacSecret    string
	redirectURL   string
	onSuccessPath string
	jwtExpiration time.Duration
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) *Config {
	var config Config

	flags.New("ClientID", "Client ID").Prefix(prefix).DocPrefix("github").StringVar(fs, &config.clientID, "", overrides)
	flags.New("ClientSecret", "Client Secret").Prefix(prefix).DocPrefix("github").StringVar(fs, &config.clientSecret, "", overrides)
	flags.New("HmacSecret", "HMAC Secret").Prefix(prefix).DocPrefix("github").StringVar(fs, &config.hmacSecret, "", overrides)
	flags.New("JwtExpiration", "JWT Expiration").Prefix(prefix).DocPrefix("github").DurationVar(fs, &config.jwtExpiration, time.Hour*24*5, overrides)
	flags.New("RedirectURL", "URL used for redirection").Prefix(prefix).DocPrefix("github").StringVar(fs, &config.redirectURL, "http://127.0.0.1/auth/github/callback", overrides)
	flags.New("OnSuccessPath", "Path for redirecting on success").Prefix(prefix).DocPrefix("github").StringVar(fs, &config.onSuccessPath, "/", overrides)

	return &config
}

func New(config *Config, cache Cache, provider Provider) Service {
	return Service{
		config: oauth2.Config{
			ClientID:     config.clientID,
			ClientSecret: config.clientSecret,
			Endpoint:     github.Endpoint,
			RedirectURL:  config.redirectURL,
			Scopes:       nil,
		},
		hmacSecret:    []byte(config.hmacSecret),
		jwtExpiration: config.jwtExpiration,
		onSuccessPath: config.onSuccessPath,

		cache:    cache,
		provider: provider,
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
		httperror.HandleError(ctx, w, httpmodel.WrapNotFound(fmt.Errorf("state not found: %w", err)))
		return
	}

	var payload State
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		httperror.HandleError(ctx, w, httpmodel.WrapNotFound(fmt.Errorf("unmarshal state: %w", err)))
		return
	}

	isRegistration := len(payload.Registration) != 0

	oauth2Token, err := s.config.Exchange(ctx, r.URL.Query().Get("code"), oauth2.VerifierOption(payload.Verifier))
	if err != nil {
		httperror.HandleError(ctx, w, httpmodel.WrapUnauthorized(fmt.Errorf("exchange token: %w", err)))
		return
	}

	if err := s.cache.Delete(ctx, state); err != nil {
		httperror.HandleError(ctx, w, httpmodel.WrapNotFound(fmt.Errorf("delete state: %w", err)))
		return
	}

	client := s.config.Client(ctx, oauth2Token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		httperror.InternalServerError(ctx, w, fmt.Errorf("get /user: %w", err))
		return
	}

	githubUser, err := httpjson.Read[model.User](resp)
	if err != nil {
		httperror.InternalServerError(ctx, w, fmt.Errorf("read /user: %w", err))
		return
	}

	var login string

	if isRegistration {
		login = payload.Registration
	} else {
		login = strconv.FormatUint(githubUser.ID, 10)
	}

	user, err := s.provider.GetGitHubUser(ctx, login)
	if err != nil {
		if errors.Is(err, model.ErrUnknownUser) {
			httperror.HandleError(ctx, w, httpmodel.WrapNotFound(fmt.Errorf("unregistered user `%s`", login)))
			return
		}

		httperror.InternalServerError(ctx, w, fmt.Errorf("get user: %w", err))
		return
	}

	token := jwt.NewWithClaims(signMethod, s.newClaim(oauth2Token, user))

	tokenString, err := token.SignedString(s.hmacSecret)
	if err != nil {
		httperror.InternalServerError(ctx, w, fmt.Errorf("sign JWT: %w", err))
		return
	}

	s.setCallbackCookie(w, cookieName, tokenString)

	if isRegistration {
		if err := s.provider.UpdateGitHubUser(ctx, user, strconv.FormatUint(githubUser.ID, 10), githubUser.Name); err != nil {
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
	<body>
		<a href="%[1]s">Continue...</a>
	</body>
</html>`, redirectPath)
}
