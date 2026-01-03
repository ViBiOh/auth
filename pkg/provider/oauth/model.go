package oauth

import (
	"github.com/ViBiOh/auth/v3/pkg/model"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

type State struct {
	Verifier     string `json:"verifier"`
	Registration string `json:"registration"`
	Redirection  string `json:"redirect"`
}

type AuthClaims struct {
	Token *oauth2.Token `json:"token"`
	jwt.RegisteredClaims
	User model.User `json:"user"`
}
