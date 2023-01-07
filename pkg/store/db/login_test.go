package db

import (
	"context"
	"errors"
	"testing"

	"github.com/ViBiOh/auth/v2/pkg/ident"
	"github.com/ViBiOh/auth/v2/pkg/mocks"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v4"
)

func TestLogin(t *testing.T) {
	t.Parallel()

	type args struct {
		login    string
		password string
	}

	cases := map[string]struct {
		args    args
		want    model.User
		wantErr error
	}{
		"simple": {
			args{
				login:    "vibioh",
				password: "secret",
			},
			model.NewUser(1, "vibioh"),
			nil,
		},
		"not found": {
			args{
				login:    "vibioh",
				password: "secret",
			},
			model.User{},
			ident.ErrInvalidCredentials,
		},
		"error": {
			args{
				login:    "vibioh",
				password: "secret",
			},
			model.User{},
			ident.ErrUnavailableService,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := App{db: mockDatabase}

			switch intention {
			case "simple":
				mockRow := mocks.NewRow(ctrl)
				mockRow.EXPECT().Scan(gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					*pointers[0].(*uint64) = 1
					*pointers[1].(*string) = "vibioh"

					return nil
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Row) error, _ string, _ ...any) error {
					return scanner(mockRow)
				}
				mockDatabase.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), "vibioh", "secret").DoAndReturn(dummyFn)

			case "not found":
				mockRow := mocks.NewRow(ctrl)
				mockRow.EXPECT().Scan(gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					return pgx.ErrNoRows
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Row) error, _ string, _ ...any) error {
					return scanner(mockRow)
				}
				mockDatabase.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), "vibioh", "secret").DoAndReturn(dummyFn)

			case "error":
				mockDatabase.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), "vibioh", "secret").Return(errors.New("timeout"))
			}

			got, gotErr := instance.Login(context.TODO(), testCase.args.login, testCase.args.password)
			failed := false

			if testCase.wantErr != nil && !errors.Is(gotErr, testCase.wantErr) {
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

	type args struct {
		user    model.User
		profile string
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"simple": {
			args{
				user:    model.NewUser(1, "vibioh"),
				profile: "admin",
			},
			true,
		},
		"error": {
			args{
				user:    model.NewUser(1, "vibioh"),
				profile: "admin",
			},
			false,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := App{db: mockDatabase}

			switch intention {
			case "simple":
				mockRow := mocks.NewRow(ctrl)
				mockRow.EXPECT().Scan(gomock.Any()).DoAndReturn(func(pointers ...any) error {
					*pointers[0].(*uint64) = 1

					return nil
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Row) error, _ string, _ ...any) error {
					return scanner(mockRow)
				}
				mockDatabase.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), uint64(1), "admin").DoAndReturn(dummyFn)
			case "error":
				mockDatabase.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), uint64(1), "admin").Return(errors.New("timeout"))
			}

			if got := instance.IsAuthorized(context.TODO(), testCase.args.user, testCase.args.profile); got != testCase.want {
				t.Errorf("IsAuthorized() = %t, want %t", got, testCase.want)
			}
		})
	}
}
