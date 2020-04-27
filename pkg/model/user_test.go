package model

import (
	"context"
	"reflect"
	"testing"
)

func TestNewUser(t *testing.T) {
	type args struct {
		id    uint64
		login string
	}

	var cases = []struct {
		intention string
		args      args
		want      User
	}{
		{
			"simple",
			args{
				id:    1,
				login: "vibioh",
			},
			User{ID: 1, Login: "vibioh"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := NewUser(tc.args.id, tc.args.login); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("NewUser() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestReadUser(t *testing.T) {
	type args struct {
		ctx context.Context
	}

	var cases = []struct {
		intention string
		args      args
		want      User
	}{
		{
			"empty",
			args{
				ctx: context.Background(),
			},
			NoneUser,
		},
		{
			"with User",
			args{
				ctx: StoreUser(context.Background(), NewUser(8000, "vibioh")),
			},
			NewUser(8000, "vibioh"),
		},
		{
			"not an User",
			args{
				ctx: context.WithValue(context.Background(), ctxUserKey, args{}),
			},
			NoneUser,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := ReadUser(tc.args.ctx); got != tc.want {
				t.Errorf("ReadUser() = %v, want %v", got, tc.want)
			}
		})
	}
}
