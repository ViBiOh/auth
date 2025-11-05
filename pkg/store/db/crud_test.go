package db

import (
	"context"
	"errors"
	"testing"

	"github.com/ViBiOh/auth/v3/pkg/mocks"
	"github.com/ViBiOh/auth/v3/pkg/model"
	"go.uber.org/mock/gomock"
)

func TestCreate(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		wantErr error
	}{
		"create": {
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
			case "create":
				mockDatabase.EXPECT().One(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			}

			got, gotErr := instance.Create(context.Background())

			failed := false

			if testCase.wantErr == nil && gotErr != nil {
				failed = true
			} else if testCase.wantErr != nil && !errors.Is(gotErr, testCase.wantErr) {
				failed = true
			} else if len(got.ID) == 0 {
				failed = true
			}

			if failed {
				t.Errorf("Create() = (`%s`), want (`%s`)", gotErr, testCase.wantErr)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()

	deletedUser := model.NewUser("")

	cases := map[string]struct {
		user    model.User
		wantErr error
	}{
		"delete": {
			deletedUser,
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
			case "delete":
				mockDatabase.EXPECT().One(gomock.Any(), gomock.Any(), deletedUser.ID).Return(nil)
			}

			gotErr := instance.Delete(context.Background(), testCase.user)

			failed := false

			if testCase.wantErr == nil && gotErr != nil {
				failed = true
			} else if testCase.wantErr != nil && !errors.Is(gotErr, testCase.wantErr) {
				failed = true
			}

			if failed {
				t.Errorf("Delete() = `%s`, want `%s`", gotErr, testCase.wantErr)
			}
		})
	}
}
