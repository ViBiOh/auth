package model

import "context"

type key int

const (
	ctxUserKey key = iota
)

type User struct {
	Name string `json:"name"`
	ID   uint64 `json:"id"`
}

func (u User) IsZero() bool {
	return u.ID == 0 && len(u.Name) == 0
}

func NewUser(id uint64, name string) User {
	return User{
		ID:   id,
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
