package memory

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/ViBiOh/auth/v3/pkg/argon"
	"github.com/ViBiOh/auth/v3/pkg/model"
)

func TestLogin(t *testing.T) {
	t.Parallel()

	argonPassword, err := argon.GenerateFromPassword("password")
	if err != nil {
		t.Errorf("generate password: %s", err)
	}

	adminUser := model.NewUser("admin")

	instance := Service{
		identifications: map[string]basicUser{
			adminUser.ID: {
				adminUser,
				[]byte(argonPassword),
			},
		},
	}

	type args struct {
		login    string
		password string
	}

	cases := map[string]struct {
		args    args
		want    model.User
		wantErr error
	}{
		"unknown": {
			args{
				login: "anonymous",
			},
			model.User{},
			model.ErrInvalidCredentials,
		},
		"invalid password": {
			args{
				login:    "admin",
				password: "admin",
			},
			model.User{},
			model.ErrInvalidCredentials,
		},
		"success": {
			args{
				login:    adminUser.ID,
				password: "password",
			},
			adminUser,
			nil,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			got, gotErr := instance.GetBasicUser(context.Background(), testCase.args.login, testCase.args.password)

			failed := false

			if testCase.wantErr == nil && gotErr != nil {
				failed = true
			} else if testCase.wantErr != nil && !errors.Is(gotErr, testCase.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, testCase.want) {
				failed = true
			}

			if failed {
				t.Errorf("Login() = (%v, `%s`), want (%v, `%s`)", got, gotErr, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestIsAuthorized(t *testing.T) {
	t.Parallel()

	adminUser := model.NewUser("root")

	instance := Service{
		authorizations: map[string][]string{
			adminUser.ID: {"admin"},
			"2":          nil,
		},
	}

	type args struct {
		user    model.User
		profile string
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"unknown": {
			args{
				user: model.NewUser("vibioh"),
			},
			true,
		},
		"no wanted profile": {
			args{
				user:    model.NewUser("vibioh"),
				profile: "",
			},
			true,
		},
		"no matching profile": {
			args{
				user:    model.NewUser("guest"),
				profile: "admin",
			},
			false,
		},
		"success": {
			args{
				user:    adminUser,
				profile: "admin",
			},
			true,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if got := instance.IsAuthorized(context.Background(), testCase.args.user, testCase.args.profile); got != testCase.want {
				t.Errorf("IsAuthorized() = %t, want %t", got, testCase.want)
			}
		})
	}
}
