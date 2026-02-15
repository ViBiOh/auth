package model

import (
	"golang.org/x/oauth2"
)

type OAuthClaim struct {
	Token *oauth2.Token `json:"token"`
	User  User          `json:"user"`
}

func (oac OAuthClaim) GetSubject() string {
	return oac.User.GetSubject()
}

type DiscordUser struct {
	Username string `json:"global_name"`
	Avatar   string `json:"avatar"`
	ID       string `json:"id"`
}

func (du DiscordUser) GetID() string {
	return du.ID
}

type GitHubUser struct {
	Login string `json:"login"`
	ID    uint64 `json:"id"`
}

func (gu GitHubUser) GetID() uint64 {
	return gu.ID
}
