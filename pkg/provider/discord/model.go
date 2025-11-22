package discord

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

type User struct {
	Username string `json:"global_name"`
	Avatar   string `json:"avatar"`
	ID       string `json:"id"`
}

type AuthClaims struct {
	Token *oauth2.Token `json:"token"`
	jwt.RegisteredClaims
	User model.User `json:"user"`
}
