package db

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/db"
)

var (
	sortKeyMatcher = regexp.MustCompile(`[A-Za-z0-9]+`)
)

func (a app) DoAtomic(ctx context.Context, action func(context.Context) error) error {
	return db.DoAtomic(ctx, a.db, action)
}

const listQuery = `
SELECT
  id,
  login,
  count(1) OVER() AS full_count
FROM
  auth.login
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

	var totalCount uint
	list := make([]model.User, 0)

	scanner := func(rows *sql.Rows) error {
		var user model.User

		if err := rows.Scan(&user.ID, &user.Login, &totalCount); err != nil {
			return err
		}

		list = append(list, user)
		return nil
	}

	err := db.List(ctx, a.db, scanner, fmt.Sprintf(listQuery, order), pageSize, offset)
	return list, totalCount, err
}

const getByIDQuery = `
SELECT
  id,
  login
FROM
  auth.login
WHERE
  id = $1
`

func (a app) Get(ctx context.Context, id uint64) (model.User, error) {
	var item model.User
	scanner := func(row *sql.Row) error {
		err := row.Scan(&item.ID, &item.Login)
		if err == sql.ErrNoRows {
			item = model.NoneUser
			return nil
		}

		return err
	}

	err := db.Get(ctx, a.db, scanner, getByIDQuery, id)
	return item, err
}

const insertQuery = `
INSERT INTO
  auth.login
(
  login,
  password
) VALUES (
  $1,
  crypt($2, gen_salt('bf',8))
) RETURNING id
`

func (a app) Create(ctx context.Context, o model.User) (uint64, error) {
	return db.Create(ctx, insertQuery, strings.ToLower(o.Login), o.Password)
}

const updateQuery = `
UPDATE
  auth.login
SET
  login = $2
WHERE
  id = $1
`

func (a app) Update(ctx context.Context, o model.User) error {
	return db.Exec(ctx, updateQuery, o.ID, strings.ToLower(o.Login))
}

const updatePasswordQuery = `
UPDATE
  auth.login
SET
  password = crypt($2, gen_salt('bf',8))
WHERE
  id = $1
`

func (a app) UpdatePassword(ctx context.Context, o model.User) error {
	return db.Exec(ctx, updatePasswordQuery, o.ID, o.Password)
}

const deleteQuery = `
DELETE FROM
  auth.login
WHERE
  id = $1
`

func (a app) Delete(ctx context.Context, o model.User) error {
	return db.Exec(ctx, deleteQuery, o.ID)
}