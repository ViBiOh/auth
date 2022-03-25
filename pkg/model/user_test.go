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

	cases := map[string]struct {
		args args
		want User
	}{
		"simple": {
			args{
				id:    1,
				login: "vibioh",
			},
			User{ID: 1, Login: "vibioh"},
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
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

	cases := map[string]struct {
		args args
		want User
	}{
		"empty": {
			args{
				ctx: context.Background(),
			},
			User{},
		},
		"with User": {
			args{
				ctx: StoreUser(context.Background(), NewUser(8000, "vibioh")),
			},
			NewUser(8000, "vibioh"),
		},
		"not an User": {
			args{
				ctx: context.WithValue(context.Background(), ctxUserKey, args{}),
			},
			User{},
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			if got := ReadUser(tc.args.ctx); got != tc.want {
				t.Errorf("ReadUser() = %v, want %v", got, tc.want)
			}
		})
	}
}
