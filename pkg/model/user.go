package model

import "strings"

// User of the app
type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	profiles string
}

// NewUser creates new user with given id, username and profiles
func NewUser(id uint, username string, profiles string) *User {
	return &User{ID: id, Username: username, profiles: profiles}
}

// HasProfile check if User has given profile
func (user *User) HasProfile(profile string) bool {
	return strings.Contains(user.profiles, profile)
}
