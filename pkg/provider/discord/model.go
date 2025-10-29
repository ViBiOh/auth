package discord

import (
	"fmt"

	"github.com/ViBiOh/auth/v3/pkg/model"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

type State struct {
	Verifier     string `json:"verifier"`
	Registration string `json:"registration"`
}

type User struct {
	Username string `json:"global_name"`
	Avatar   string `json:"avatar"`
	ID       string `json:"id"`
}

func (u User) Image() string {
	return fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.webp", u.ID, u.Avatar)
}

type AuthClaims struct {
	Token *oauth2.Token `json:"token"`
	jwt.RegisteredClaims
	User model.User `json:"user"`
}
