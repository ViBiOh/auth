package model

import "strings"

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

// HasProfile check if User has given profile
func (user *User) HasProfile(profile string) bool {
	return strings.Contains(user.profiles, profile)
}
