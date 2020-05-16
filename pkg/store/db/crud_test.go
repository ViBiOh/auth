package db

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/db"
)

func TestList(t *testing.T) {
	type args struct {
		page     uint
		pageSize uint
		sortKey  string
		sortAsc  bool
	}

	var cases = []struct {
		intention string
		args      args
		expectSQL string
		want      []model.User
		wantCount uint
		wantErr   error
	}{
		{
			"simple",
			args{
				page:     1,
				pageSize: 20,
			},
			"SELECT id, login, .+ AS full_count FROM login ORDER BY creation_date DESC",
			[]model.User{
				model.NewUser(1, "vibioh"),
				model.NewUser(2, "guest"),
			},
			2,
			nil,
		},
		{
			"timeout",
			args{
				page:     1,
				pageSize: 20,
			},
			"SELECT id, login, .+ AS full_count FROM login ORDER BY creation_date DESC",
			nil,
			0,
			sqlmock.ErrCancelled,
		},
		{
			"order",
			args{
				page:     1,
				pageSize: 20,
				sortKey:  "login",
				sortAsc:  true,
			},
			"SELECT id, login, .+ AS full_count FROM login ORDER BY login",
			[]model.User{
				model.NewUser(1, "vibioh"),
				model.NewUser(2, "guest"),
			},
			2,
			nil,
		},
		{
			"descending",
			args{
				page:     1,
				pageSize: 20,
				sortKey:  "login",
				sortAsc:  false,
			},
			"SELECT id, login, .+ AS full_count FROM login ORDER BY login DESC",
			[]model.User{
				model.NewUser(1, "vibioh"),
				model.NewUser(2, "guest"),
			},
			2,
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			mockDb, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer mockDb.Close()

			expectedQuery := mock.ExpectQuery(tc.expectSQL).WithArgs(20, 0).WillReturnRows(sqlmock.NewRows([]string{"id", "login", "full_count"}).AddRow(1, "vibioh", 2).AddRow(2, "guest", 2))

			if tc.intention == "timeout" {
				savedSQLTimeout := db.SQLTimeout
				db.SQLTimeout = time.Second
				defer func() {
					db.SQLTimeout = savedSQLTimeout
				}()

				expectedQuery.WillDelayFor(db.SQLTimeout * 2)
			}

			got, gotCount, gotErr := app{db: mockDb}.List(context.Background(), tc.args.page, tc.args.pageSize, tc.args.sortKey, tc.args.sortAsc)
			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			} else if gotCount != tc.wantCount {
				failed = true
			}

			if failed {
				t.Errorf("List() = (%+v, %d, `%s`), want (%+v, %d, `%s`)", got, gotCount, gotErr, tc.want, tc.wantCount, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}

func TestGet(t *testing.T) {
	type args struct {
		id uint64
	}

	var cases = []struct {
		intention string
		args      args
		want      model.User
		wantErr   error
	}{
		{
			"create",
			args{
				id: 1,
			},
			model.NewUser(1, "vibioh"),
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			mockDb, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer mockDb.Close()

			mock.ExpectQuery("SELECT id, login FROM login").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id", "login"}).AddRow(1, "vibioh"))

			got, gotErr := app{db: mockDb}.Get(context.Background(), tc.args.id)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("Get() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
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
		want      uint64
		wantErr   error
	}{
		{
			"create",
			args{
				o: model.User{
					Login:    "vibioh",
					Password: "secret",
				},
			},
			1,
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			mockDb, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer mockDb.Close()

			ctx := context.Background()
			mock.ExpectBegin()
			tx, err := mockDb.Begin()
			if err != nil {
				t.Errorf("unable to create tx: %s", err)
			}
			ctx = db.StoreTx(ctx, tx)

			mock.ExpectQuery("INSERT INTO login").WithArgs("vibioh", "secret").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

			got, gotErr := app{db: mockDb}.Create(ctx, tc.args.o)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if got != tc.want {
				failed = true
			}

			if failed {
				t.Errorf("Create() = (%d, `%s`), want (%d, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
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
		wantErr   error
	}{
		{
			"update",
			args{
				o: model.User{
					ID:    1,
					Login: "vibioh",
				},
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			mockDb, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer mockDb.Close()

			ctx := context.Background()
			mock.ExpectBegin()
			tx, err := mockDb.Begin()
			if err != nil {
				t.Errorf("unable to create tx: %s", err)
			}
			ctx = db.StoreTx(ctx, tx)

			mock.ExpectExec("UPDATE login SET login").WithArgs(1, "vibioh").WillReturnResult(sqlmock.NewResult(0, 1))

			gotErr := app{db: mockDb}.Update(ctx, tc.args.o)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && !errors.Is(gotErr, tc.wantErr) {
				failed = true
			}

			if failed {
				t.Errorf("Update() = `%s`, want `%s`", gotErr, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}

func TestUpdatePassword(t *testing.T) {
	type args struct {
		o model.User
	}

	var cases = []struct {
		intention string
		args      args
		wantErr   error
	}{
		{
			"update",
			args{
				o: model.User{
					ID:       1,
					Password: "secret",
				},
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			mockDb, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer mockDb.Close()

			ctx := context.Background()
			mock.ExpectBegin()
			tx, err := mockDb.Begin()
			if err != nil {
				t.Errorf("unable to create tx: %s", err)
			}
			ctx = db.StoreTx(ctx, tx)

			mock.ExpectExec("UPDATE login SET password").WithArgs(1, "secret").WillReturnResult(sqlmock.NewResult(0, 1))

			gotErr := app{db: mockDb}.UpdatePassword(ctx, tc.args.o)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && !errors.Is(gotErr, tc.wantErr) {
				failed = true
			}

			if failed {
				t.Errorf("UpdatePassword() = `%s`, want `%s`", gotErr, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
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
		wantErr   error
	}{
		{
			"update",
			args{
				o: model.User{
					ID: 1,
				},
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			mockDb, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer mockDb.Close()

			ctx := context.Background()
			mock.ExpectBegin()
			tx, err := mockDb.Begin()
			if err != nil {
				t.Errorf("unable to create tx: %s", err)
			}
			ctx = db.StoreTx(ctx, tx)

			mock.ExpectExec("DELETE FROM login").WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))

			gotErr := app{db: mockDb}.Delete(ctx, tc.args.o)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && !errors.Is(gotErr, tc.wantErr) {
				failed = true
			}

			if failed {
				t.Errorf("Delete() = `%s`, want `%s`", gotErr, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}
