package db

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/ViBiOh/auth/v2/pkg/mocks"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/jackc/pgx/v5"
	"go.uber.org/mock/gomock"
)

func TestGet(t *testing.T) {
	t.Parallel()

	type args struct {
		id uint64
	}

	cases := map[string]struct {
		args    args
		want    model.User
		wantErr error
	}{
		"create": {
			args{
				id: 1,
			},
			model.NewUser(1, "vibioh"),
			nil,
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
			case "create":
				mockRow := mocks.NewRow(ctrl)
				mockRow.EXPECT().Scan(gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					*pointers[0].(*uint64) = 1
					*pointers[1].(*string) = "vibioh"

					return nil
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Row) error, _ string, _ ...any) error {
					return scanner(mockRow)
				}
				mockDatabase.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), uint64(1)).DoAndReturn(dummyFn)
			}

			got, gotErr := instance.Get(context.TODO(), testCase.args.id)

			failed := false

			if testCase.wantErr == nil && gotErr != nil {
				failed = true
			} else if testCase.wantErr != nil && !errors.Is(gotErr, testCase.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, testCase.want) {
				failed = true
			}

			if failed {
				t.Errorf("Get() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	t.Parallel()

	type args struct {
		o model.User
	}

	cases := map[string]struct {
		args    args
		want    uint64
		wantErr error
	}{
		"create": {
			args{
				o: model.User{
					Login:    "ViBiOh",
					Password: "secret",
				},
			},
			1,
			nil,
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
			case "create":
				mockDatabase.EXPECT().Create(gomock.Any(), gomock.Any(), "vibioh", "secret").Return(uint64(1), nil)
			}

			got, gotErr := instance.Create(context.TODO(), testCase.args.o)

			failed := false

			if testCase.wantErr == nil && gotErr != nil {
				failed = true
			} else if testCase.wantErr != nil && !errors.Is(gotErr, testCase.wantErr) {
				failed = true
			} else if got != testCase.want {
				failed = true
			}

			if failed {
				t.Errorf("Create() = (%d, `%s`), want (%d, `%s`)", got, gotErr, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	type args struct {
		o model.User
	}

	cases := map[string]struct {
		args    args
		wantErr error
	}{
		"update": {
			args{
				o: model.User{
					ID:    1,
					Login: "ViBiOh",
				},
			},
			nil,
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
			case "update":
				mockDatabase.EXPECT().One(gomock.Any(), gomock.Any(), uint64(1), "vibioh").Return(nil)
			}

			gotErr := instance.Update(context.TODO(), testCase.args.o)

			failed := false

			if testCase.wantErr == nil && gotErr != nil {
				failed = true
			} else if testCase.wantErr != nil && !errors.Is(gotErr, testCase.wantErr) {
				failed = true
			}

			if failed {
				t.Errorf("Update() = `%s`, want `%s`", gotErr, testCase.wantErr)
			}
		})
	}
}

func TestUpdatePassword(t *testing.T) {
	t.Parallel()

	type args struct {
		o model.User
	}

	cases := map[string]struct {
		args    args
		wantErr error
	}{
		"update": {
			args{
				o: model.User{
					ID:       1,
					Password: "secret",
				},
			},
			nil,
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
			case "update":
				mockDatabase.EXPECT().One(gomock.Any(), gomock.Any(), uint64(1), "secret").Return(nil)
			}

			gotErr := instance.UpdatePassword(context.TODO(), testCase.args.o)

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

func TestDelete(t *testing.T) {
	t.Parallel()

	type args struct {
		o model.User
	}

	cases := map[string]struct {
		args    args
		wantErr error
	}{
		"delete": {
			args{
				o: model.User{
					ID: 1,
				},
			},
			nil,
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
			case "delete":
				mockDatabase.EXPECT().One(gomock.Any(), gomock.Any(), uint64(1)).Return(nil)
			}

			gotErr := instance.Delete(context.TODO(), testCase.args.o)

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
