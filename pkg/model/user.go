package model

// NoneUser is a dummy user
var NoneUser User

// User of the app
type User struct {
	ID       uint64 `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password,omitempty"`
}

// NewUser creates new user with given id, login and profiles
func NewUser(id uint64, login string) User {
	return User{
		ID:    id,
		Login: login,
	}
}
