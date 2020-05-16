package service

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/ViBiOh/auth/v2/pkg/model"
)

type testStore struct{}

func (ts testStore) DoAtomic(ctx context.Context, action func(ctx context.Context) error) error {
	return action(ctx)
}

func (ts testStore) List(ctx context.Context, page, pageSize uint, sortKey string, sortAsc bool) ([]model.User, uint, error) {
	if sortKey == "error" {
		return nil, 0, errors.New("invalid sort key")
	}
	return []model.User{
		model.NewUser(1, "admin"),
		model.NewUser(2, "guest"),
	}, 2, nil
}

func (ts testStore) Get(ctx context.Context, id uint64) (model.User, error) {
	if id == 8000 {
		return model.NoneUser, errors.New("unable to connect")
	}

	if id == 2 {
		return model.NoneUser, nil
	}

	return model.NewUser(id, "admin"), nil
}

func (ts testStore) Create(ctx context.Context, o model.User) (uint64, error) {
	if o.ID != 0 {
		return 0, errors.New("invalid id")
	}

	return 1, nil
}

func (ts testStore) Update(ctx context.Context, o model.User) error {
	if o.ID == 0 {
		return errors.New("unable to connect")
	}

	return nil
}

func (ts testStore) Delete(ctx context.Context, o model.User) error {
	if o.ID == 0 {
		return errors.New("unable to connect")
	}

	return nil
}

func (ts testStore) IsAuthorized(ctx context.Context, user model.User, profile string) bool {
	return user.Login == "admin"
}

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
		args      args
		want      []model.User
		wantCount uint
		wantErr   error
	}{
		{
			"no context",
			args{
				ctx: context.Background(),
			},
			nil,
			0,
			ErrUnauthorized,
		},
		{
			"not admin",
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(1, "guest")),
			},
			nil,
			0,
			ErrForbidden,
		},
		{
			"error on list",
			args{
				ctx:     model.StoreUser(context.Background(), model.NewUser(1, "admin")),
				sortKey: "error",
			},
			nil,
			0,
			errors.New("unable to list: invalid sort key"),
		},
		{
			"valid",
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
			got, gotCount, gotErr := New(testStore{}, testStore{}).List(tc.args.ctx, tc.args.page, tc.args.pageSize, tc.args.sortKey, tc.args.sortAsc, tc.args.filters)

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
		args      args
		want      model.User
		wantErr   error
	}{
		{
			"no context",
			args{
				ctx: context.Background(),
			},
			model.NoneUser,
			ErrUnauthorized,
		},
		{
			"not self",
			args{
				id:  1,
				ctx: model.StoreUser(context.Background(), model.NewUser(2, "guest")),
			},
			model.NoneUser,
			ErrForbidden,
		},
		{
			"error on get",
			args{
				id:  8000,
				ctx: model.StoreUser(context.Background(), model.NewUser(1, "admin")),
			},
			model.NoneUser,
			errors.New("unable to get: unable to connect"),
		},
		{
			"not found",
			args{
				id:  2,
				ctx: model.StoreUser(context.Background(), model.NewUser(2, "guest")),
			},
			model.NoneUser,
			ErrNotFound,
		},
		{
			"found",
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
			got, gotErr := New(testStore{}, testStore{}).Get(tc.args.ctx, tc.args.id)

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
		args      args
		want      model.User
		wantErr   error
	}{
		{
			"error on create",
			args{
				o: model.NewUser(1, "admin"),
			},
			model.NoneUser,
			errors.New("unable to create: invalid id"),
		},
		{
			"success",
			args{
				o: model.NewUser(0, "admin"),
			},
			model.NewUser(1, "admin"),
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			got, gotErr := New(testStore{}, testStore{}).Create(context.Background(), tc.args.o)

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
		args      args
		want      model.User
		wantErr   error
	}{
		{
			"error on update",
			args{
				o: model.NewUser(0, "admin"),
			},
			model.NewUser(0, "admin"),
			errors.New("unable to update: unable to connect"),
		},
		{
			"success",
			args{
				o: model.NewUser(1, "admin"),
			},
			model.NewUser(1, "admin"),
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			got, gotErr := New(testStore{}, testStore{}).Update(context.Background(), tc.args.o)

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
		args      args
		want      model.User
		wantErr   error
	}{
		{
			"error on delete",
			args{
				o: model.NewUser(0, "admin"),
			},
			model.NewUser(0, "admin"),
			errors.New("unable to delete: unable to connect"),
		},
		{
			"success",
			args{
				o: model.NewUser(1, "admin"),
			},
			model.NewUser(1, "admin"),
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			gotErr := New(testStore{}, testStore{}).Delete(context.Background(), tc.args.o)

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
		args      args
		wantErr   error
	}{
		{
			"empty",
			args{
				ctx: context.Background(),
			},
			errors.New("you must be an admin for deleting"),
		},
		{
			"create empty",
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
			args{
				ctx: context.Background(),
				new: model.NewUser(0, "guest"),
			},
			errors.New("password is required"),
		},
		{
			"create valid",
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
			args{
				ctx: context.Background(),
				old: model.NewUser(2, "guest"),
				new: model.NewUser(2, ""),
			},
			errors.New("you must be logged in for interacting, you're not authorized to interact with other user, login is required"),
		},
		{
			"update forbidden",
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(1, "guest")),
				old: model.NewUser(2, "guest"),
				new: model.NewUser(2, ""),
			},
			errors.New("you're not authorized to interact with other user, login is required"),
		},
		{
			"update empty login",
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(2, "guest")),
				old: model.NewUser(2, "guest"),
				new: model.NewUser(2, ""),
			},
			errors.New("login is required"),
		},
		{
			"update valid",
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(2, "guest")),
				old: model.NewUser(2, "guest"),
				new: model.NewUser(2, "guest_new"),
			},
			nil,
		},
		{
			"update as admin",
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(1, "admin")),
				old: model.NewUser(2, "guest"),
				new: model.NewUser(2, "guest_new"),
			},
			nil,
		},
		{
			"delete unauthorized",
			args{
				ctx: context.Background(),
				old: model.NewUser(2, "guest"),
			},
			errors.New("you must be logged in for interacting, you must be an admin for deleting"),
		},
		{
			"delete forbidden",
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(1, "guest")),
				old: model.NewUser(2, "guest"),
			},
			errors.New("you must be an admin for deleting"),
		},
		{
			"delete self",
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(2, "guest")),
				old: model.NewUser(2, "guest"),
			},
			errors.New("you must be an admin for deleting"),
		},
		{
			"delete admin",
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(1, "admin")),
				old: model.NewUser(2, "guest"),
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			gotErr := New(testStore{}, testStore{}).Check(tc.args.ctx, tc.args.old, tc.args.new)

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
