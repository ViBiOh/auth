package store

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"

	"github.com/ViBiOh/auth/v2/pkg/auth"
	"github.com/ViBiOh/auth/v2/pkg/ident/basic"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/auth/v2/pkg/store"
	"github.com/ViBiOh/httputils/v3/pkg/db"
)

var (
	_ auth.Provider     = app{}
	_ basic.Provider    = app{}
	_ store.UserStorage = app{}

	sortKeyMatcher = regexp.MustCompile(`[A-Za-z0-9]+`)
)

// App of package
type App interface {
	List(ctx context.Context, page, pageSize uint, sortKey string, sortAsc bool) ([]model.User, uint, error)
	Get(ctx context.Context, id uint64) (model.User, error)
	Create(ctx context.Context, o model.User) (uint64, error)
	Update(ctx context.Context, o model.User) error
	Delete(ctx context.Context, o model.User) error

	Login(ctx context.Context, login, password string) (model.User, error)
	IsAuthorized(ctx context.Context, user model.User, profile string) bool
}

type app struct {
	db *sql.DB
}

// New creates new App from dependencies
func New(db *sql.DB) App {
	return app{
		db: db,
	}
}

func scanUser(row db.RowScanner) (model.User, error) {
	var user model.User

	if err := row.Scan(&user.ID, &user.Login); err != nil {
		if err == sql.ErrNoRows {
			return model.NoneUser, nil
		}

		return model.NoneUser, err
	}

	return user, nil
}

func scanUsers(rows *sql.Rows) ([]model.User, uint, error) {
	var totalCount uint
	list := make([]model.User, 0)

	for rows.Next() {
		var user model.User

		if err := rows.Scan(&user.ID, &user.Login, &totalCount); err != nil {
			return nil, 0, err
		}

		list = append(list, user)
	}

	return list, totalCount, nil
}

const listQuery = `
SELECT
  id,
  login,
  count(id) OVER() AS full_count
FROM
  login
ORDER BY %s
LIMIT $1
OFFSET $2
`

func (a app) List(ctx context.Context, page, pageSize uint, sortKey string, sortAsc bool) ([]model.User, uint, error) {
	order := "creation_date DESC"

	if sortKeyMatcher.MatchString(sortKey) {
		order = sortKey

		if !sortAsc {
			order += " DESC"
		}
	}

	offset := (page - 1) * pageSize

	ctx, cancel := context.WithTimeout(ctx, db.SQLTimeout)
	defer cancel()

	rows, err := a.db.QueryContext(ctx, fmt.Sprintf(listQuery, order), pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	defer func() {
		err = db.RowsClose(rows, err)
	}()

	return scanUsers(rows)
}

const getByIDQuery = `
SELECT
  id,
  login
FROM
  login
WHERE
  id = $1
`

func (a app) Get(ctx context.Context, id uint64) (model.User, error) {
	return scanUser(db.GetRow(ctx, a.db, getByIDQuery, id))
}

const insertQuery = `
INSERT INTO
  login
(
  login,
  password
) VALUES (
  $1,
  crypt($2, gen_salt('bf',8))
) RETURNING id
`

func (a app) Create(ctx context.Context, o model.User) (uint64, error) {
	return db.Create(ctx, a.db, insertQuery, o.Login, o.Password)
}

const updateQuery = `
UPDATE
  login
SET
  login = $2
WHERE
  id = $1
`

func (a app) Update(ctx context.Context, o model.User) error {
	return db.Exec(ctx, a.db, updateQuery, o.ID, o.Login)
}

const deleteQuery = `
DELETE FROM
  login
WHERE
  id = $1
`

func (a app) Delete(ctx context.Context, o model.User) error {
	return db.Exec(ctx, a.db, deleteQuery, o.ID)
}
