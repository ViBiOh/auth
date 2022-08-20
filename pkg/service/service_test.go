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
	"github.com/golang/mock/gomock"
)

func TestGet(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx context.Context
		id  uint64
	}

	cases := map[string]struct {
		instance App
		args     args
		want     model.User
		wantErr  error
	}{
		"no context": {
			App{},
			args{
				ctx: context.Background(),
			},
			model.User{},
			httpModel.ErrUnauthorized,
		},
		"not self": {
			App{},
			args{
				id:  1,
				ctx: model.StoreUser(context.Background(), model.NewUser(2, "guest")),
			},
			model.User{},
			httpModel.ErrForbidden,
		},
		"error on get": {
			App{},
			args{
				id:  8000,
				ctx: model.StoreUser(context.Background(), model.NewUser(1, "admin")),
			},
			model.User{},
			errors.New("get: failed"),
		},
		"not found": {
			App{},
			args{
				id:  2,
				ctx: model.StoreUser(context.Background(), model.NewUser(2, "guest")),
			},
			model.User{},
			httpModel.ErrNotFound,
		},
		"found": {
			App{},
			args{
				id:  1,
				ctx: model.StoreUser(context.Background(), model.NewUser(2, "admin")),
			},
			model.NewUser(1, "admin"),
			nil,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authStorage := mocks.NewStorage(ctrl)
			authProvider := mocks.NewProvider(ctrl)

			switch intention {
			case "not self":
				authProvider.EXPECT().IsAuthorized(gomock.Any(), gomock.Any(), gomock.Any()).Return(false)
			case "error on get":
				authStorage.EXPECT().Get(gomock.Any(), gomock.Any()).Return(model.User{}, errors.New("failed"))
				authProvider.EXPECT().IsAuthorized(gomock.Any(), gomock.Any(), gomock.Any()).Return(true)
			case "not found":
				authStorage.EXPECT().Get(gomock.Any(), gomock.Any()).Return(model.User{}, nil)
			case "found":
				authStorage.EXPECT().Get(gomock.Any(), gomock.Any()).Return(model.NewUser(1, "admin"), nil)
				authProvider.EXPECT().IsAuthorized(gomock.Any(), gomock.Any(), gomock.Any()).Return(true)
			}

			testCase.instance.storeApp = authStorage
			testCase.instance.authApp = authProvider

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
		instance App
		args     args
		want     model.User
		wantErr  error
	}{
		"error on create": {
			App{},
			args{
				o: model.NewUser(1, "admin"),
			},
			model.User{},
			errors.New("create: failed"),
		},
		"success": {
			App{},
			args{
				o: model.NewUser(0, "admin"),
			},
			model.NewUser(1, "admin"),
			nil,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authStorage := mocks.NewStorage(ctrl)

			testCase.instance.storeApp = authStorage

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
		instance App
		args     args
		want     model.User
		wantErr  error
	}{
		"error on update": {
			App{},
			args{
				o: model.NewUser(1, "admin"),
			},
			model.NewUser(1, "admin"),
			errors.New("update: failed"),
		},
		"success": {
			App{},
			args{
				o: model.NewUser(1, "admin"),
			},
			model.NewUser(1, "admin"),
			nil,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authStorage := mocks.NewStorage(ctrl)

			testCase.instance.storeApp = authStorage

			switch intention {
			case "error on update":
				authStorage.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.New("failed"))
			case "success":
				authStorage.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
			}

			got, gotErr := testCase.instance.Update(context.Background(), testCase.args.o)

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
				t.Errorf("Update() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, testCase.want, testCase.wantErr)
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
		instance App
		args     args
		want     model.User
		wantErr  error
	}{
		"error on delete": {
			App{},
			args{
				o: model.NewUser(0, "admin"),
			},
			model.NewUser(0, "admin"),
			errors.New("delete: failed"),
		},
		"success": {
			App{},
			args{
				o: model.NewUser(1, "admin"),
			},
			model.NewUser(1, "admin"),
			nil,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authStorage := mocks.NewStorage(ctrl)

			testCase.instance.storeApp = authStorage

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

func TestCheck(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx context.Context
		old model.User
		new model.User
	}

	cases := map[string]struct {
		instance App
		args     args
		wantErr  error
	}{
		"empty": {
			App{},
			args{
				ctx: context.Background(),
			},
			errors.New("you must be an admin for deleting"),
		},
		"create empty": {
			App{},
			args{
				ctx: context.Background(),
				new: model.User{
					ID: 1,
				},
			},
			errors.New("login is required, password is required"),
		},
		"create without password": {
			App{},
			args{
				ctx: context.Background(),
				new: model.NewUser(0, "guest"),
			},
			errors.New("password is required"),
		},
		"create valid": {
			App{},
			args{
				ctx: context.Background(),
				new: model.User{
					Login:    "guest",
					Password: "secret",
				},
			},
			nil,
		},
		"update unauthorized": {
			App{},
			args{
				ctx: context.Background(),
				old: model.NewUser(2, "guest"),
				new: model.NewUser(2, ""),
			},
			errors.New("you must be logged in for interacting, you're not authorized to interact with other user, login is required"),
		},
		"update forbidden": {
			App{},
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(1, "guest")),
				old: model.NewUser(2, "guest"),
				new: model.NewUser(2, ""),
			},
			errors.New("you're not authorized to interact with other user, login is required"),
		},
		"update empty login": {
			App{},
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(2, "guest")),
				old: model.NewUser(2, "guest"),
				new: model.NewUser(2, ""),
			},
			errors.New("login is required"),
		},
		"update valid": {
			App{},
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(2, "guest")),
				old: model.NewUser(2, "guest"),
				new: model.NewUser(2, "guest_new"),
			},
			nil,
		},
		"update as admin": {
			App{},
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(1, "admin")),
				old: model.NewUser(2, "guest"),
				new: model.NewUser(2, "guest_new"),
			},
			nil,
		},
		"delete unauthorized": {
			App{},
			args{
				ctx: context.Background(),
				old: model.NewUser(2, "guest"),
			},
			errors.New("you must be logged in for interacting, you must be an admin for deleting"),
		},
		"delete forbidden": {
			App{},
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(1, "guest")),
				old: model.NewUser(2, "guest"),
			},
			errors.New("you must be an admin for deleting"),
		},
		"delete self": {
			App{},
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(2, "guest")),
				old: model.NewUser(2, "guest"),
			},
			errors.New("you must be an admin for deleting"),
		},
		"delete admin": {
			App{},
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(1, "admin")),
				old: model.NewUser(2, "guest"),
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

			authStorage := mocks.NewStorage(ctrl)
			authProvider := mocks.NewProvider(ctrl)

			testCase.instance.storeApp = authStorage
			testCase.instance.authApp = authProvider

			switch intention {
			case "empty":
				authProvider.EXPECT().IsAuthorized(gomock.Any(), gomock.Any(), gomock.Any()).Return(false)
			case "update unauthorized":
				authProvider.EXPECT().IsAuthorized(gomock.Any(), gomock.Any(), gomock.Any()).Return(false)
			case "update forbidden":
				authProvider.EXPECT().IsAuthorized(gomock.Any(), gomock.Any(), gomock.Any()).Return(false)
			case "delete unauthorized":
				authProvider.EXPECT().IsAuthorized(gomock.Any(), gomock.Any(), gomock.Any()).Return(false)
			case "delete forbidden":
				authProvider.EXPECT().IsAuthorized(gomock.Any(), gomock.Any(), gomock.Any()).Return(false)
			case "update as admin":
				authProvider.EXPECT().IsAuthorized(gomock.Any(), gomock.Any(), gomock.Any()).Return(true)
			case "delete self":
				authProvider.EXPECT().IsAuthorized(gomock.Any(), gomock.Any(), gomock.Any()).Return(false)
			case "delete admin":
				authProvider.EXPECT().IsAuthorized(gomock.Any(), gomock.Any(), gomock.Any()).Return(true)
			}

			gotErr := testCase.instance.Check(testCase.args.ctx, testCase.args.old, testCase.args.new)

			failed := false

			if testCase.wantErr == nil && gotErr != nil {
				failed = true
			} else if testCase.wantErr != nil && gotErr == nil {
				failed = true
			} else if testCase.wantErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()) {
				failed = true
			}

			if failed {
				t.Errorf("Check() = `%s`, want `%s`", gotErr, testCase.wantErr)
			}
		})
	}
}
