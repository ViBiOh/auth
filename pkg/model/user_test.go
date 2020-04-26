package model

import (
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
