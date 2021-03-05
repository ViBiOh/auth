package service

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/ViBiOh/auth/v2/pkg/auth/authtest"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/auth/v2/pkg/store/storetest"
	httpModel "github.com/ViBiOh/httputils/v4/pkg/model"
)

func TestList(t *testing.T) {
	type args struct {
		ctx      context.Context
		page     uint
		pageSize uint
		sortKey  string
		sortAsc  bool
		filters  map[string][]string
	}

	var cases = []struct {
		intention string
		instance  App
		args      args
		want      []model.User
		wantCount uint
		wantErr   error
	}{
		{
			"no context",
			New(storetest.New(), authtest.New()),
			args{
				ctx: context.Background(),
			},
			nil,
			0,
			httpModel.ErrUnauthorized,
		},
		{
			"not admin",
			New(storetest.New(), authtest.New().SetIsAuthorized(false)),
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(1, "guest")),
			},
			nil,
			0,
			httpModel.ErrForbidden,
		},
		{
			"error on list",
			New(storetest.New().SetList(nil, 0, errors.New("failed")), authtest.New().SetIsAuthorized(true)),
			args{
				ctx:     model.StoreUser(context.Background(), model.NewUser(1, "admin")),
				sortKey: "error",
			},
			nil,
			0,
			errors.New("unable to list: failed"),
		},
		{
			"valid",
			New(storetest.New().SetList([]model.User{
				model.NewUser(1, "admin"),
				model.NewUser(2, "guest"),
			}, 2, nil), authtest.New().SetIsAuthorized(true)),
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(1, "admin")),
			},
			[]model.User{
				model.NewUser(1, "admin"),
				model.NewUser(2, "guest"),
			},
			2,
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			got, gotCount, gotErr := tc.instance.List(tc.args.ctx, tc.args.page, tc.args.pageSize, tc.args.sortKey, tc.args.sortAsc, tc.args.filters)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			} else if gotCount != tc.wantCount {
				failed = true
			}

			if failed {
				t.Errorf("List() = (%+v, %d, `%s`), want (%+v, %d, `%s`)", got, gotCount, gotErr, tc.want, tc.wantCount, tc.wantErr)
			}
		})
	}
}

func TestGet(t *testing.T) {
	type args struct {
		ctx context.Context
		id  uint64
	}

	var cases = []struct {
		intention string
		instance  App
		args      args
		want      model.User
		wantErr   error
	}{
		{
			"no context",
			New(storetest.New(), authtest.New()),
			args{
				ctx: context.Background(),
			},
			model.NoneUser,
			httpModel.ErrUnauthorized,
		},
		{
			"not self",
			New(storetest.New(), authtest.New()),
			args{
				id:  1,
				ctx: model.StoreUser(context.Background(), model.NewUser(2, "guest")),
			},
			model.NoneUser,
			httpModel.ErrForbidden,
		},
		{
			"error on get",
			New(storetest.New().SetGet(model.NoneUser, errors.New("failed")), authtest.New().SetIsAuthorized(true)),
			args{
				id:  8000,
				ctx: model.StoreUser(context.Background(), model.NewUser(1, "admin")),
			},
			model.NoneUser,
			errors.New("unable to get: failed"),
		},
		{
			"not found",
			New(storetest.New().SetGet(model.NoneUser, nil), authtest.New().SetIsAuthorized(true)),
			args{
				id:  2,
				ctx: model.StoreUser(context.Background(), model.NewUser(2, "guest")),
			},
			model.NoneUser,
			httpModel.ErrNotFound,
		},
		{
			"found",
			New(storetest.New().SetGet(model.NewUser(1, "admin"), nil), authtest.New().SetIsAuthorized(true)),
			args{
				id:  1,
				ctx: model.StoreUser(context.Background(), model.NewUser(2, "admin")),
			},
			model.NewUser(1, "admin"),
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			got, gotErr := tc.instance.Get(tc.args.ctx, tc.args.id)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("Get() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type args struct {
		o model.User
	}

	var cases = []struct {
		intention string
		instance  App
		args      args
		want      model.User
		wantErr   error
	}{
		{
			"error on create",
			New(storetest.New().SetCreate(0, errors.New("failed")), nil),
			args{
				o: model.NewUser(1, "admin"),
			},
			model.NoneUser,
			errors.New("unable to create: failed"),
		},
		{
			"success",
			New(storetest.New().SetCreate(1, nil), nil),
			args{
				o: model.NewUser(0, "admin"),
			},
			model.NewUser(1, "admin"),
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			got, gotErr := tc.instance.Create(context.Background(), tc.args.o)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("Create() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type args struct {
		o model.User
	}

	var cases = []struct {
		intention string
		instance  App
		args      args
		want      model.User
		wantErr   error
	}{
		{
			"error on update",
			New(storetest.New().SetUpdate(errors.New("failed")), nil),
			args{
				o: model.NewUser(1, "admin"),
			},
			model.NewUser(1, "admin"),
			errors.New("unable to update: failed"),
		},
		{
			"success",
			New(storetest.New().SetUpdate(nil), nil),
			args{
				o: model.NewUser(1, "admin"),
			},
			model.NewUser(1, "admin"),
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			got, gotErr := tc.instance.Update(context.Background(), tc.args.o)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("Update() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		o model.User
	}

	var cases = []struct {
		intention string
		instance  App
		args      args
		want      model.User
		wantErr   error
	}{
		{
			"error on delete",
			New(storetest.New().SetDelete(errors.New("failed")), nil),
			args{
				o: model.NewUser(0, "admin"),
			},
			model.NewUser(0, "admin"),
			errors.New("unable to delete: failed"),
		},
		{
			"success",
			New(storetest.New().SetDelete(nil), nil),
			args{
				o: model.NewUser(1, "admin"),
			},
			model.NewUser(1, "admin"),
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			gotErr := tc.instance.Delete(context.Background(), tc.args.o)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			}

			if failed {
				t.Errorf("Delete() = `%s`, want `%s`", gotErr, tc.wantErr)
			}
		})
	}
}

func TestCheck(t *testing.T) {
	type args struct {
		ctx context.Context
		old model.User
		new model.User
	}

	var cases = []struct {
		intention string
		instance  App
		args      args
		wantErr   error
	}{
		{
			"empty",
			New(storetest.New(), authtest.New()),
			args{
				ctx: context.Background(),
			},
			errors.New("you must be an admin for deleting"),
		},
		{
			"create empty",
			New(storetest.New(), authtest.New()),
			args{
				ctx: context.Background(),
				new: model.User{
					ID: 1,
				},
			},
			errors.New("login is required, password is required"),
		},
		{
			"create without password",
			New(storetest.New(), authtest.New()),
			args{
				ctx: context.Background(),
				new: model.NewUser(0, "guest"),
			},
			errors.New("password is required"),
		},
		{
			"create valid",
			New(storetest.New(), authtest.New()),
			args{
				ctx: context.Background(),
				new: model.User{
					Login:    "guest",
					Password: "secret",
				},
			},
			nil,
		},
		{
			"update unauthorized",
			New(storetest.New(), authtest.New()),
			args{
				ctx: context.Background(),
				old: model.NewUser(2, "guest"),
				new: model.NewUser(2, ""),
			},
			errors.New("you must be logged in for interacting, you're not authorized to interact with other user, login is required"),
		},
		{
			"update forbidden",
			New(storetest.New(), authtest.New()),
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(1, "guest")),
				old: model.NewUser(2, "guest"),
				new: model.NewUser(2, ""),
			},
			errors.New("you're not authorized to interact with other user, login is required"),
		},
		{
			"update empty login",
			New(storetest.New(), authtest.New()),
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(2, "guest")),
				old: model.NewUser(2, "guest"),
				new: model.NewUser(2, ""),
			},
			errors.New("login is required"),
		},
		{
			"update valid",
			New(storetest.New(), authtest.New()),
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(2, "guest")),
				old: model.NewUser(2, "guest"),
				new: model.NewUser(2, "guest_new"),
			},
			nil,
		},
		{
			"update as admin",
			New(storetest.New(), authtest.New().SetIsAuthorized(true)),
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(1, "admin")),
				old: model.NewUser(2, "guest"),
				new: model.NewUser(2, "guest_new"),
			},
			nil,
		},
		{
			"delete unauthorized",
			New(storetest.New(), authtest.New()),
			args{
				ctx: context.Background(),
				old: model.NewUser(2, "guest"),
			},
			errors.New("you must be logged in for interacting, you must be an admin for deleting"),
		},
		{
			"delete forbidden",
			New(storetest.New(), authtest.New()),
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(1, "guest")),
				old: model.NewUser(2, "guest"),
			},
			errors.New("you must be an admin for deleting"),
		},
		{
			"delete self",
			New(storetest.New(), authtest.New()),
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(2, "guest")),
				old: model.NewUser(2, "guest"),
			},
			errors.New("you must be an admin for deleting"),
		},
		{
			"delete admin",
			New(storetest.New(), authtest.New().SetIsAuthorized(true)),
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(1, "admin")),
				old: model.NewUser(2, "guest"),
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			gotErr := tc.instance.Check(tc.args.ctx, tc.args.old, tc.args.new)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			}

			if failed {
				t.Errorf("Check() = `%s`, want `%s`", gotErr, tc.wantErr)
			}
		})
	}
}
