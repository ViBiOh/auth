package github

import (
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

type State struct {
	Verifier     string `json:"verifier"`
	Registration string `json:"registration"`
}

type AuthClaims struct {
	Token *oauth2.Token `json:"token"`
	jwt.RegisteredClaims
	User model.User `json:"user"`
}
