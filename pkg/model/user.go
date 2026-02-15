package model

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/ViBiOh/httputils/v4/pkg/id"
)

type key int

const (
	ctxUserKey key = iota
)

type User struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Image string   `json:"image"`
	Kind  UserKind `json:"kind"`
}

func NewUser(name string) User {
	return User{
		ID:   id.New(),
		Name: name,
	}
}

func (u User) GetSubject() string {
	return u.Name
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

//go:generate stringer -type=UserKind
type UserKind int

const (
	Invite UserKind = iota
	GitHub
	Discord
	Basic
)

var ErrUnknownUserKind = errors.New("unknown UserKind")

var UserKinds = sync.OnceValue(func() []UserKind {
	count := len(_UserKind_index) - 1
	values := make([]UserKind, count)

	for i := range count {
		values[i] = UserKind(i)
	}

	return values
})

func ParseUserKind(value string) (UserKind, error) {
	var previous uint8

	for i := 1; i < len(_UserKind_index); i++ {
		current := _UserKind_index[i]

		if strings.EqualFold(_UserKind_name[previous:current], value) {
			return UserKind(i - 1), nil
		}

		previous = current
	}

	return Invite, fmt.Errorf("parse `%s`: %w", value, ErrUnknownUserKind)
}

func (e UserKind) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(e.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (e *UserKind) UnmarshalJSON(b []byte) error {
	var strValue string

	if err := json.Unmarshal(b, &strValue); err != nil {
		return fmt.Errorf("unmarshal UserKind: %w", err)
	}

	value, err := ParseUserKind(strValue)
	if err != nil {
		return err
	}

	*e = value

	return nil
}
