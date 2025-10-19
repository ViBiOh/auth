package github

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ViBiOh/auth/v3/pkg/model"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

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
			ID:        strconv.FormatInt(int64(user.ID), 10),
		},
	}
}

func (s Service) setCallbackCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   int(s.jwtExpiration.Seconds()),
		Path:     "/",
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}
