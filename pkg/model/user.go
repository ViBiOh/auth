package model

import "context"

type key int

const (
	ctxUserKey key = iota
)

// User of the app
type User struct {
	Login    string `json:"login"`
	Password string `json:"password,omitempty"`
	ID       uint64 `json:"id"`
}

// IsZero check if instance is valued or not
func (u User) IsZero() bool {
	return u.ID == 0 && len(u.Login) == 0
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
func ReadUser(ctx context.Context) (output User) {
	rawUser := ctx.Value(ctxUserKey)
	if rawUser == nil {
		return User{}
	}

	output, _ = rawUser.(User)
	return
}
