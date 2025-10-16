package service

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/ViBiOh/auth/v2/pkg/mocks"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"go.uber.org/mock/gomock"
)

func TestCreate(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		instance Service
		want     model.User
		wantErr  error
	}{
		"error on create": {
			Service{},
			model.User{},
			errors.New("failed"),
		},
		"success": {
			Service{},
			model.NewUser(1, "admin"),
			nil,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			authStorage := mocks.NewMockStorage(ctrl)
			testCase.instance.storage = authStorage

			switch intention {
			case "error on create":
				authStorage.EXPECT().Create(gomock.Any()).Return(model.User{}, errors.New("failed"))
			case "success":
				authStorage.EXPECT().Create(gomock.Any()).Return(model.User{ID: 1, Name: "admin"}, nil)
			}

			got, gotErr := testCase.instance.Create(context.Background())

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

			authStorage := mocks.NewMockStorage(ctrl)
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
