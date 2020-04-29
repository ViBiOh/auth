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
  count(1) OVER() AS full_count
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
	var item model.User
	scanner := func(row db.RowScanner) error {
		err := row.Scan(&item.ID, &item.Login)
		if err == sql.ErrNoRows {
			item = model.NoneUser
			return nil
		}

		return err
	}

	err := db.GetRow(ctx, a.db, scanner, getByIDQuery, id)
	return item, err
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
	return db.Create(ctx, a.db, insertQuery, strings.ToLower(o.Login), o.Password)
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
	return db.Exec(ctx, a.db, updateQuery, o.ID, strings.ToLower(o.Login))
}

const updatePasswordQuery = `
UPDATE
  login
SET
  password = crypt($2, gen_salt('bf',8))
WHERE
  id = $1
`

func (a app) UpdatePassword(ctx context.Context, o model.User) error {
	return db.Exec(ctx, a.db, updatePasswordQuery, o.ID, o.Password)
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
