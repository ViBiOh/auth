package db

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ViBiOh/auth/v2/pkg/ident"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/db"
)

func TestLogin(t *testing.T) {
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
			"simple",
			args{
				login:    "vibioh",
				password: "secret",
			},
			model.NewUser(1, "vibioh"),
			nil,
		},
		{
			"not found",
			args{
				login:    "vibioh",
				password: "secret",
			},
			model.NoneUser,
			ident.ErrInvalidCredentials,
		},
		{
			"timeout",
			args{
				login:    "vibioh",
				password: "secret",
			},
			model.NoneUser,
			ident.ErrUnavailableService,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			mockDb, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer mockDb.Close()

			expectedQuery := mock.ExpectQuery("SELECT id, login FROM auth.login WHERE login = .+ AND password = crypt(.+, password)").WithArgs("vibioh", "secret")

			if tc.intention != "not found" {
				expectedQuery.WillReturnRows(sqlmock.NewRows([]string{"id", "login"}).AddRow(1, "vibioh"))
			} else {
				expectedQuery.WillReturnRows(sqlmock.NewRows([]string{"id", "login"}))
			}

			if tc.intention == "timeout" {
				savedSQLTimeout := db.SQLTimeout
				db.SQLTimeout = time.Second
				defer func() {
					db.SQLTimeout = savedSQLTimeout
				}()

				expectedQuery.WillDelayFor(db.SQLTimeout * 2)
			}

			got, gotErr := New(mockDb).Login(context.Background(), tc.args.login, tc.args.password)
			failed := false

			if tc.wantErr != nil && !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if got != tc.want {
				failed = true
			}

			if failed {
				t.Errorf("Login() = (%v, `%s`), want (%v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}

func TestIsAuthorized(t *testing.T) {
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
			"simple",
			args{
				user:    model.NewUser(1, "vibioh"),
				profile: "admin",
			},
			true,
		},
		{
			"timeout",
			args{
				user:    model.NewUser(1, "vibioh"),
				profile: "admin",
			},
			false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			mockDb, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer mockDb.Close()

			expectedQuery := mock.ExpectQuery("SELECT p.id FROM auth.profile p, auth.login_profile lp WHERE p.name = .+ AND lp.profile_id = p.id AND lp.login_id = .+").WithArgs(1, "admin").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

			if tc.intention == "timeout" {
				savedSQLTimeout := db.SQLTimeout
				db.SQLTimeout = time.Second
				defer func() {
					db.SQLTimeout = savedSQLTimeout
				}()

				expectedQuery.WillDelayFor(db.SQLTimeout * 2)
			}

			if got := New(mockDb).IsAuthorized(context.Background(), tc.args.user, tc.args.profile); got != tc.want {
				t.Errorf("IsAuthorized() = %t, want %t", got, tc.want)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}
