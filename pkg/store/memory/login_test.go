package memory

import (
	"context"
	"errors"
	"testing"

	"github.com/ViBiOh/auth/v2/pkg/ident"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"golang.org/x/crypto/bcrypt"
)

func TestLogin(t *testing.T) {
	t.Parallel()

	passwordValue, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		t.Errorf("generate password: %s", err)
	}

	instance := App{
		ident: map[string]basicUser{
			"admin": {
				model.NewUser(1, "admin"),
				passwordValue,
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
			ident.ErrInvalidCredentials,
		},
		"invalid password": {
			args{
				login:    "admin",
				password: "admin",
			},
			model.User{},
			ident.ErrInvalidCredentials,
		},
		"success": {
			args{
				login:    "admin",
				password: "password",
			},
			model.NewUser(1, "admin"),
			nil,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			got, gotErr := instance.Login(context.TODO(), testCase.args.login, testCase.args.password)

			failed := false

			if testCase.wantErr == nil && gotErr != nil {
				failed = true
			} else if testCase.wantErr != nil && !errors.Is(gotErr, testCase.wantErr) {
				failed = true
			} else if got != testCase.want {
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

	instance := App{
		auth: map[uint64][]string{
			1: {"admin"},
			2: nil,
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
				user: model.NewUser(8000, "vibioh"),
			},
			false,
		},
		"no wanted profile": {
			args{
				user:    model.NewUser(1, "vibioh"),
				profile: "",
			},
			true,
		},
		"no matching profile": {
			args{
				user:    model.NewUser(2, "guest"),
				profile: "admin",
			},
			false,
		},
		"success": {
			args{
				user:    model.NewUser(1, "vibioh"),
				profile: "admin",
			},
			true,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if got := instance.IsAuthorized(context.TODO(), testCase.args.user, testCase.args.profile); got != testCase.want {
				t.Errorf("IsAuthorized() = %t, want %t", got, testCase.want)
			}
		})
	}
}
