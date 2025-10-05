package db

import (
	"context"
	"errors"
	"testing"

	"github.com/ViBiOh/auth/v2/pkg/mocks"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/jackc/pgx/v5"
	"go.uber.org/mock/gomock"
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
			model.ErrInvalidCredentials,
		},
		"error": {
			args{
				login:    "vibioh",
				password: "secret",
			},
			model.User{},
			model.ErrUnavailableService,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := Service{db: mockDatabase}

			switch intention {
			case "simple":
				mockRow := mocks.NewRow(ctrl)
				mockRow.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					*pointers[0].(*uint64) = 1
					*pointers[1].(*string) = "vibioh"
					*pointers[2].(*string) = "$argon2id$v=19$m=7168,t=5,p=1$Fh3xnr+CV5ymbbx9hnfWQsEZOzSc0nI$/NU9AeurqbuHYx75qNFNDJxsUDqevR2eJnQSLNw8OMA"

					return nil
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Row) error, _ string, _ ...any) error {
					return scanner(mockRow)
				}
				mockDatabase.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), "vibioh").DoAndReturn(dummyFn)

			case "not found":
				mockRow := mocks.NewRow(ctrl)
				mockRow.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					return pgx.ErrNoRows
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Row) error, _ string, _ ...any) error {
					return scanner(mockRow)
				}
				mockDatabase.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), "vibioh").DoAndReturn(dummyFn)

			case "error":
				mockDatabase.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), "vibioh").Return(errors.New("timeout"))
			}

			got, gotErr := instance.Login(context.Background(), nil, testCase.args.login, testCase.args.password)
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

func TestUpdatePassword(t *testing.T) {
	t.Parallel()

	type args struct {
		o        model.User
		password string
	}

	cases := map[string]struct {
		args    args
		wantErr error
	}{
		"update": {
			args{
				o: model.User{
					ID: 1,
				},
				password: "secret",
			},
			nil,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := Service{db: mockDatabase}

			switch intention {
			case "update":
				mockDatabase.EXPECT().One(gomock.Any(), gomock.Any(), uint64(1), gomock.Any()).Return(nil)
			}

			gotErr := instance.UpdatePassword(context.Background(), testCase.args.o, testCase.args.password)

			failed := false

			if testCase.wantErr == nil && gotErr != nil {
				failed = true
			} else if testCase.wantErr != nil && !errors.Is(gotErr, testCase.wantErr) {
				failed = true
			}

			if failed {
				t.Errorf("UpdatePassword() = `%s`, want `%s`", gotErr, testCase.wantErr)
			}
		})
	}
}
