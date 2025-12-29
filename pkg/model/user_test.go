package model

import (
	"context"
	"reflect"
	"testing"
)

func TestReadUser(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx context.Context
	}

	user := NewUser("vibioh")

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
				ctx: StoreUser(context.Background(), user),
			},
			user,
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

			if got := ReadUser(testCase.args.ctx); !reflect.DeepEqual(got, testCase.want) {
				t.Errorf("ReadUser() = %v, want %v", got, testCase.want)
			}
		})
	}
}
