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
	Username string `json:"username"`
	profiles string
}

// NewUser creates new user with given id, username and profiles
func NewUser(id uint64, username, profiles string) User {
	return User{
		ID:       id,
		Username: username,
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
