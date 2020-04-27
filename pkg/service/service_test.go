package service

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/crud"
)

type testStore struct{}

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

func TestUnmarshal(t *testing.T) {
	type args struct {
		data        []byte
		contentType string
	}

	var cases = []struct {
		intention string
		args      args
		want      model.User
		wantErr   error
	}{
		{
			"invalid",
			args{
				data: []byte("{\"id\": 1,\"login\": \"vibioh\""),
			},
			model.NoneUser,
			errors.New("unexpected end of JSON input"),
		},
		{
			"valid",
			args{
				data: []byte("{\"id\": 1,\"login\": \"vibioh\"}"),
			},
			model.NewUser(1, "vibioh"),
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			got, gotErr := New(testStore{}, testStore{}).Unmarshal(tc.args.data, tc.args.contentType)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			} else if got != tc.want {
				failed = true
			}

			if failed {
				t.Errorf("Unmarshal() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
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
		want      []interface{}
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
			crud.ErrUnauthorized,
		},
		{
			"not admin",
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(1, "guest")),
			},
			nil,
			0,
			crud.ErrForbidden,
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
			[]interface{}{
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
		want      interface{}
		wantErr   error
	}{
		{
			"no context",
			args{
				ctx: context.Background(),
			},
			nil,
			crud.ErrUnauthorized,
		},
		{
			"not self",
			args{
				id:  1,
				ctx: model.StoreUser(context.Background(), model.NewUser(2, "guest")),
			},
			nil,
			crud.ErrForbidden,
		},
		{
			"error on get",
			args{
				id:  8000,
				ctx: model.StoreUser(context.Background(), model.NewUser(1, "admin")),
			},
			nil,
			errors.New("unable to get: unable to connect"),
		},
		{
			"not found",
			args{
				id:  2,
				ctx: model.StoreUser(context.Background(), model.NewUser(2, "guest")),
			},
			nil,
			crud.ErrNotFound,
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
		o interface{}
	}

	var cases = []struct {
		intention string
		args      args
		want      interface{}
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
		o interface{}
	}

	var cases = []struct {
		intention string
		args      args
		want      interface{}
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
		o interface{}
	}

	var cases = []struct {
		intention string
		args      args
		want      interface{}
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
		old interface{}
		new interface{}
	}

	var cases = []struct {
		intention string
		args      args
		want      []crud.Error
	}{
		{
			"empty",
			args{
				ctx: context.Background(),
			},
			[]crud.Error{
				crud.NewError("context", "you must be an admin for deleting"),
			},
		},
		{
			"create empty",
			args{
				ctx: context.Background(),
				new: model.NoneUser,
			},
			[]crud.Error{
				crud.NewError("login", "login is required"),
				crud.NewError("password", "password is required"),
			},
		},
		{
			"create without password",
			args{
				ctx: context.Background(),
				new: model.NewUser(0, "guest"),
			},
			[]crud.Error{
				crud.NewError("password", "password is required"),
			},
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
			[]crud.Error{},
		},
		{
			"update unauthorized",
			args{
				ctx: context.Background(),
				old: model.NewUser(2, "guest"),
				new: model.NewUser(2, ""),
			},
			[]crud.Error{
				crud.NewError("context", "you must be logged in for interacting"),
				crud.NewError("context", "you're not authorized to interact with other user"),
				crud.NewError("login", "login is required"),
			},
		},
		{
			"update forbidden",
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(1, "guest")),
				old: model.NewUser(2, "guest"),
				new: model.NewUser(2, ""),
			},
			[]crud.Error{
				crud.NewError("context", "you're not authorized to interact with other user"),
				crud.NewError("login", "login is required"),
			},
		},
		{
			"update empty login",
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(2, "guest")),
				old: model.NewUser(2, "guest"),
				new: model.NewUser(2, ""),
			},
			[]crud.Error{
				crud.NewError("login", "login is required"),
			},
		},
		{
			"update valid",
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(2, "guest")),
				old: model.NewUser(2, "guest"),
				new: model.NewUser(2, "guest_new"),
			},
			[]crud.Error{},
		},
		{
			"update as admin",
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(1, "admin")),
				old: model.NewUser(2, "guest"),
				new: model.NewUser(2, "guest_new"),
			},
			[]crud.Error{},
		},
		{
			"delete unauthorized",
			args{
				ctx: context.Background(),
				old: model.NewUser(2, "guest"),
			},
			[]crud.Error{
				crud.NewError("context", "you must be logged in for interacting"),
				crud.NewError("context", "you must be an admin for deleting"),
			},
		},
		{
			"delete forbidden",
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(1, "guest")),
				old: model.NewUser(2, "guest"),
			},
			[]crud.Error{
				crud.NewError("context", "you must be an admin for deleting"),
			},
		},
		{
			"delete self",
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(2, "guest")),
				old: model.NewUser(2, "guest"),
			},
			[]crud.Error{
				crud.NewError("context", "you must be an admin for deleting"),
			},
		},
		{
			"delete admin",
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(1, "admin")),
				old: model.NewUser(2, "guest"),
			},
			[]crud.Error{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := New(testStore{}, testStore{}).Check(tc.args.ctx, tc.args.old, tc.args.new); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Check() = %+v, want %+v", got, tc.want)
			}
		})
	}
}
