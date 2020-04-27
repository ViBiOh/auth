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
	passwordValue, err := bcrypt.GenerateFromPassword([]byte("password"), 12)
	if err != nil {
		t.Errorf("unable to generate password: %s", err)
	}

	instance := app{
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

	var cases = []struct {
		intention string
		args      args
		want      model.User
		wantErr   error
	}{
		{
			"unknown",
			args{
				login: "anonymous",
			},
			model.NoneUser,
			ident.ErrInvalidCredentials,
		},
		{
			"invalid password",
			args{
				login:    "admin",
				password: "admin",
			},
			model.NoneUser,
			ident.ErrInvalidCredentials,
		},
		{
			"success",
			args{
				login:    "admin",
				password: "password",
			},
			model.NewUser(1, "admin"),
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			got, gotErr := instance.Login(context.Background(), tc.args.login, tc.args.password)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if got != tc.want {
				failed = true
			}

			if failed {
				t.Errorf("Login() = (%v, `%s`), want (%v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}

func TestIsAuthorized(t *testing.T) {
	instance := app{
		auth: map[uint64][]string{
			1: {"admin"},
			2: nil,
		},
	}

	type args struct {
		user    model.User
		profile string
	}

	var cases = []struct {
		intention string
		args      args
		want      bool
	}{
		{
			"unknown",
			args{
				user: model.NewUser(8000, "vibioh"),
			},
			false,
		},
		{
			"no wanted profile",
			args{
				user:    model.NewUser(1, "vibioh"),
				profile: "",
			},
			true,
		},
		{
			"no matching profile",
			args{
				user:    model.NewUser(2, "guest"),
				profile: "admin",
			},
			false,
		},
		{
			"success",
			args{
				user:    model.NewUser(1, "vibioh"),
				profile: "admin",
			},
			true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := instance.IsAuthorized(context.Background(), tc.args.user, tc.args.profile); got != tc.want {
				t.Errorf("IsAuthorized() = %t, want %t", got, tc.want)
			}
		})
	}
}
