package model

import (
	"context"
	"reflect"
	"testing"
)

func TestNewUser(t *testing.T) {
	t.Parallel()

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

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if got := NewUser(testCase.args.id, testCase.args.login); !reflect.DeepEqual(got, testCase.want) {
				t.Errorf("NewUser() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestReadUser(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx context.Context
	}

	cases := map[string]struct {
		args args
		want User
	}{
		"empty": {
			args{
				ctx: context.TODO(),
			},
			User{},
		},
		"with User": {
			args{
				ctx: StoreUser(context.TODO(), NewUser(8000, "vibioh")),
			},
			NewUser(8000, "vibioh"),
		},
		"not an User": {
			args{
				ctx: context.WithValue(context.TODO(), ctxUserKey, args{}),
			},
			User{},
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if got := ReadUser(testCase.args.ctx); got != testCase.want {
				t.Errorf("ReadUser() = %v, want %v", got, testCase.want)
			}
		})
	}
}
