package github

import (
	"strconv"

	"github.com/ViBiOh/auth/v3/pkg/model"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

type State struct {
	Verifier     string `json:"verifier"`
	Registration string `json:"registration"`
}

type User struct {
	Login string `json:"login"`
	ID    uint64 `json:"id"`
}

func (u User) Image() string {
	return "https://avatars.githubusercontent.com/u/" + strconv.FormatUint(u.ID, 10)
}

type AuthClaims struct {
	Token *oauth2.Token `json:"token"`
	jwt.RegisteredClaims
	User model.User `json:"user"`
}
