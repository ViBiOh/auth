package github

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ViBiOh/auth/v2/pkg/ident"
	"github.com/ViBiOh/auth/v2/pkg/middleware"
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
	signMethod       = jwt.SigningMethodHS256
	signValidMethods = []string{signMethod.Alg()}
)

type Cache interface {
	Load(ctx context.Context, key string) ([]byte, error)
	Store(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
}

type Service struct {
	config        oauth2.Config
	cache         Cache
	onSuccessPath string
	hmacSecret    []byte
	jwtExpiration time.Duration
}

var _ ident.Provider = Service{}

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

func New(config *Config, cache Cache) Service {
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
		cache:         cache,
	}
}

func (s Service) GetUser(ctx context.Context, r *http.Request) (model.User, error) {
	auth, err := r.Cookie(cookieName)
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

func (s Service) OnError(w http.ResponseWriter, r *http.Request, err error) {
	state := id.New()
	verifier := oauth2.GenerateVerifier()

	if err := s.cache.Store(r.Context(), verifierCacheKey+state, verifier, time.Minute*5); err != nil {
		httperror.InternalServerError(r.Context(), w, fmt.Errorf("save state: %w", err))
		return
	}

	http.Redirect(w, r, s.config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier)), http.StatusFound)
}

func (s Service) Callback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	state := verifierCacheKey + r.URL.Query().Get("state")

	verifier, err := s.cache.Load(ctx, state)
	if err != nil {
		httperror.HandleError(ctx, w, httpmodel.WrapNotFound(fmt.Errorf("state not found: %w", err)))
		return
	}

	oauth2Token, err := s.config.Exchange(ctx, r.URL.Query().Get("code"), oauth2.VerifierOption(string(verifier)))
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

	user, err := httpjson.Read[model.User](resp)
	if err != nil {
		httperror.InternalServerError(ctx, w, fmt.Errorf("read /user: %w", err))
		return
	}

	token := jwt.NewWithClaims(signMethod, s.newClaim(oauth2Token, user))

	tokenString, err := token.SignedString(s.hmacSecret)
	if err != nil {
		httperror.InternalServerError(ctx, w, fmt.Errorf("sign JWT: %w", err))
		return
	}

	s.setCallbackCookie(w, cookieName, tokenString)

	http.Redirect(w, r, s.onSuccessPath, http.StatusFound)
}

func (s Service) jwtKeyFunc(_ *jwt.Token) (any, error) {
	return s.hmacSecret, nil
}

func (s Service) newClaim(token *oauth2.Token, user model.User) AuthClaims {
	now := time.Now()

	return AuthClaims{
		User:  user,
		Token: token,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.jwtExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "auth",
			Subject:   user.Login,
			ID:        strconv.FormatInt(int64(user.ID), 10),
		},
	}
}

func (s Service) setCallbackCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   int(s.jwtExpiration.Seconds()),
		Secure:   false,
		HttpOnly: true,
	})
}
