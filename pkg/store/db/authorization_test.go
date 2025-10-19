package db

import (
	"context"
	"errors"
	"testing"

	"github.com/ViBiOh/auth/v3/pkg/mocks"
	"github.com/ViBiOh/auth/v3/pkg/model"
	"github.com/jackc/pgx/v5"
	"go.uber.org/mock/gomock"
)

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
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := Service{db: mockDatabase}

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

			if got := instance.IsAuthorized(context.Background(), testCase.args.user, testCase.args.profile); got != testCase.want {
				t.Errorf("IsAuthorized() = %t, want %t", got, testCase.want)
			}
		})
	}
}
