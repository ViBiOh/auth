package model

import (
	"context"
	"reflect"
	"testing"
)

func TestNewUser(t *testing.T) {
	t.Parallel()

	type args struct {
		id   uint64
		name string
	}

	cases := map[string]struct {
		args args
		want User
	}{
		"simple": {
			args{
				id:   1,
				name: "vibioh",
			},
			User{ID: 1, Name: "vibioh"},
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if got := NewUser(testCase.args.id, testCase.args.name); !reflect.DeepEqual(got, testCase.want) {
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

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if got := ReadUser(testCase.args.ctx); got != testCase.want {
				t.Errorf("ReadUser() = %v, want %v", got, testCase.want)
			}
		})
	}
}
