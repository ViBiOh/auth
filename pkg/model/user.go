package model

import (
	"strings"

	"github.com/ViBiOh/httputils/pkg/crud"
)

var _ crud.Item = &User{}

// User of the app
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	profiles string
}

// NewUser creates new user with given id, username and profiles
func NewUser(id, username, email, profiles string) *User {
	return &User{ID: id, Username: username, Email: email, profiles: profiles}
}

// SetID defines new ID
func (u *User) SetID(id string) {
	u.ID = id
}

// HasProfile check if User has given profile
func (u *User) HasProfile(profile string) bool {
	return strings.Contains(u.profiles, profile)
}
