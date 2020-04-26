package model

import "context"

type key int

const (
	ctxUserKey key = iota
)

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

// StoreUser stores given User in context
func StoreUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, ctxUserKey, user)
}

// ReadUser retrieves user from context
func ReadUser(ctx context.Context) User {
	rawUser := ctx.Value(ctxUserKey)
	if rawUser == nil {
		return NoneUser
	}

	if user, ok := rawUser.(User); ok {
		return user
	}

	return NoneUser
}
