package model

import "context"

type key int

const (
	ctxUserKey key = iota
)

type User struct {
	Login string `json:"login"`
	ID    uint64 `json:"id"`
}

func (u User) IsZero() bool {
	return u.ID == 0 && len(u.Login) == 0
}

func NewUser(id uint64, login string) User {
	return User{
		ID:    id,
		Login: login,
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
