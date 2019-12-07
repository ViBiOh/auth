package model

import (
	"github.com/ViBiOh/httputils/v3/pkg/crud"
)

var _ crud.Item = &User{}

// NoneUser is a dummy user
var NoneUser User

// User of the app
type User struct {
	ID       uint64 `json:"id"`
	Login    string `json:"login"`
	Password string `json:"-"`
}

// NewUser creates new user with given id, login and profiles
func NewUser(id uint64, login string) User {
	return User{
		ID:    id,
		Login: login,
	}
}

// SetID defines new ID
func (u *User) SetID(id uint64) {
	u.ID = id
}

// SetPassword defines new password
func (u *User) SetPassword(password string) *User {
	u.Password = password

	return u
}
