package model

import (
	"context"

	"github.com/ViBiOh/httputils/v4/pkg/id"
)

type key int

const (
	ctxUserKey key = iota
)

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (u User) IsZero() bool {
	return len(u.ID) == 0 && len(u.Name) == 0
}

func NewUser(name string) User {
	return User{
		ID:   id.New(),
		Name: name,
	}
}

func StoreUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, ctxUserKey, user)
}

func ReadUser(ctx context.Context) (output User) {
	rawUser := ctx.Value(ctxUserKey)
	if rawUser == nil {
		return User{}
	}

	output, _ = rawUser.(User)
	return output
}
