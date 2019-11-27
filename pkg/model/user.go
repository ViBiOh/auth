package model

import (
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/crud"
)

var _ crud.Item = &User{}

// NoneUser is a dummy user
var NoneUser User

// User of the app
type User struct {
	ID       uint64 `json:"id"`
	Login    string `json:"login"`
	profiles string
}

// NewUser creates new user with given id, login and profiles
func NewUser(id uint64, login, profiles string) User {
	return User{
		ID:       id,
		Login:    login,
		profiles: profiles,
	}
}

// SetID defines new ID
func (u *User) SetID(id uint64) {
	u.ID = id
}

// HasProfile check if User has given profile
func (u *User) HasProfile(profile string) bool {
	return strings.Contains(u.profiles, profile)
}
