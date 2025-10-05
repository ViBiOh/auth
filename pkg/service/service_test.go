package service

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/ViBiOh/auth/v2/pkg/mocks"
	"github.com/ViBiOh/auth/v2/pkg/model"
	httpModel "github.com/ViBiOh/httputils/v4/pkg/model"
	"go.uber.org/mock/gomock"
)

func TestGet(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx context.Context
		id  uint64
	}

	cases := map[string]struct {
		instance Service
		args     args
		want     model.User
		wantErr  error
	}{
		"error on get": {
			Service{},
			args{
				id:  8000,
				ctx: model.StoreUser(context.Background(), model.NewUser(1, "admin")),
			},
			model.User{},
			errors.New("get: failed"),
		},
		"not found": {
			Service{},
			args{
				id:  2,
				ctx: model.StoreUser(context.Background(), model.NewUser(2, "guest")),
			},
			model.User{},
			httpModel.ErrNotFound,
		},
		"found": {
			Service{},
			args{
				id:  1,
				ctx: model.StoreUser(context.Background(), model.NewUser(2, "admin")),
			},
			model.NewUser(1, "admin"),
			nil,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			authStorage := mocks.NewUpdatableStorage(ctrl)

			switch intention {
			case "error on get":
				authStorage.EXPECT().Get(gomock.Any(), gomock.Any()).Return(model.User{}, errors.New("failed"))

			case "not found":
				authStorage.EXPECT().Get(gomock.Any(), gomock.Any()).Return(model.User{}, nil)

			case "found":
				authStorage.EXPECT().Get(gomock.Any(), gomock.Any()).Return(model.NewUser(1, "admin"), nil)
			}

			testCase.instance.storage = authStorage

			got, gotErr := testCase.instance.Get(testCase.args.ctx, testCase.args.id)

			failed := false

			if testCase.wantErr == nil && gotErr != nil {
				failed = true
			} else if testCase.wantErr != nil && gotErr == nil {
				failed = true
			} else if testCase.wantErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()) {
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
		instance Service
		args     args
		want     model.User
		wantErr  error
	}{
		"error on create": {
			Service{},
			args{
				o: model.NewUser(1, "admin"),
			},
			model.User{},
			errors.New("create: failed"),
		},
		"success": {
			Service{},
			args{
				o: model.NewUser(0, "admin"),
			},
			model.NewUser(1, "admin"),
			nil,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			authStorage := mocks.NewUpdatableStorage(ctrl)
			testCase.instance.storage = authStorage

			switch intention {
			case "error on create":
				authStorage.EXPECT().Create(gomock.Any(), gomock.Any()).Return(uint64(0), errors.New("failed"))
			case "success":
				authStorage.EXPECT().Create(gomock.Any(), gomock.Any()).Return(uint64(1), nil)
			}

			got, gotErr := testCase.instance.Create(context.Background(), testCase.args.o)

			failed := false

			if testCase.wantErr == nil && gotErr != nil {
				failed = true
			} else if testCase.wantErr != nil && gotErr == nil {
				failed = true
			} else if testCase.wantErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()) {
				failed = true
			} else if !reflect.DeepEqual(got, testCase.want) {
				failed = true
			}

			if failed {
				t.Errorf("Create() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, testCase.want, testCase.wantErr)
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
		instance Service
		args     args
		wantErr  error
	}{
		"error on update": {
			Service{},
			args{
				o: model.NewUser(1, "admin"),
			},
			errors.New("failed"),
		},
		"success": {
			Service{},
			args{
				o: model.NewUser(1, "admin"),
			},
			nil,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			authStorage := mocks.NewUpdatableStorage(ctrl)
			testCase.instance.storage = authStorage

			switch intention {
			case "error on update":
				authStorage.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.New("failed"))
			case "success":
				authStorage.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
			}

			gotErr := testCase.instance.Update(context.Background(), testCase.args.o)

			failed := false

			if testCase.wantErr == nil && gotErr != nil {
				failed = true
			} else if testCase.wantErr != nil && gotErr == nil {
				failed = true
			} else if testCase.wantErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()) {
				failed = true
			}

			if failed {
				t.Errorf("Update() = (`%s`), want (`%s`)", gotErr, testCase.wantErr)
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
		instance Service
		args     args
		wantErr  error
	}{
		"error on delete": {
			Service{},
			args{
				o: model.NewUser(0, "admin"),
			},
			errors.New("failed"),
		},
		"success": {
			Service{},
			args{
				o: model.NewUser(1, "admin"),
			},
			nil,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			authStorage := mocks.NewUpdatableStorage(ctrl)
			testCase.instance.storage = authStorage

			switch intention {
			case "error on delete":
				authStorage.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(errors.New("failed"))
			case "success":
				authStorage.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil)
			}

			gotErr := testCase.instance.Delete(context.Background(), testCase.args.o)

			failed := false

			if testCase.wantErr == nil && gotErr != nil {
				failed = true
			} else if testCase.wantErr != nil && gotErr == nil {
				failed = true
			} else if testCase.wantErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()) {
				failed = true
			}

			if failed {
				t.Errorf("Delete() = `%s`, want `%s`", gotErr, testCase.wantErr)
			}
		})
	}
}
