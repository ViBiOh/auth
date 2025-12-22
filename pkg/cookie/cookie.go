package cookie

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ViBiOh/auth/v3/pkg/model"
	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/httperror"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

var (
	signMethod       = jwt.SigningMethodHS256
	signValidMethods = []string{signMethod.Alg()}
)

type AuthClaims struct {
	Token *oauth2.Token `json:"token"`
	jwt.RegisteredClaims
	User model.User `json:"user"`
}

type Service struct {
	hmacSecret    []byte
	jwtExpiration time.Duration
	devMode       bool
}

type Config struct {
	hmacSecret    string
	jwtExpiration time.Duration
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) *Config {
	var config Config

	flags.New("HmacSecret", "HMAC Secret").Prefix(prefix).DocPrefix("cookie").StringVar(fs, &config.hmacSecret, "", overrides)
	flags.New("JwtExpiration", "JWT Expiration").Prefix(prefix).DocPrefix("cookie").DurationVar(fs, &config.jwtExpiration, time.Hour*24*5, overrides)

	return &config
}

func New(config *Config) Service {
	return Service{
		hmacSecret:    []byte(config.hmacSecret),
		jwtExpiration: config.jwtExpiration,
		devMode:       os.Getenv("ENV") == "dev",
	}
}

func (s Service) Get(r *http.Request, name string) (AuthClaims, error) {
	var claim AuthClaims

	auth, err := r.Cookie(name)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return claim, model.ErrMalformedContent
		}

		return claim, fmt.Errorf("get auth cookie: %w", err)
	}

	if _, err = jwt.ParseWithClaims(auth.Value, &claim, s.jwtKeyFunc, jwt.WithValidMethods(signValidMethods)); err != nil {
		return claim, fmt.Errorf("parse JWT: %w", err)
	}

	return claim, nil
}

func (s Service) Set(ctx context.Context, w http.ResponseWriter, oauth2Token *oauth2.Token, user model.User, name string) bool {
	token := jwt.NewWithClaims(signMethod, s.newClaim(oauth2Token, user))

	tokenString, err := token.SignedString(s.hmacSecret)
	if err != nil {
		httperror.InternalServerError(ctx, w, fmt.Errorf("sign JWT: %w", err))
		return false
	}

	s.setCallbackCookie(w, name, tokenString)
	return true
}

func (s Service) Clear(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		Secure:   !s.devMode,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
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
			Subject:   user.Name,
			ID:        user.ID,
		},
	}
}

func (s Service) setCallbackCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   int(s.jwtExpiration.Seconds()),
		Path:     "/",
		Secure:   !s.devMode,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}
