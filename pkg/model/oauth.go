package model

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
