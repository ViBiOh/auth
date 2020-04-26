package service

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"time"

	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/db"
)

var (
	sortKeyMatcher = regexp.MustCompile(`[A-Za-z0-9]+`)
	sqlTimeout     = time.Second * 5
)

// RowScanner describes scan ability of a row
type RowScanner interface {
	Scan(...interface{}) error
}

func scanUser(row RowScanner) (model.User, error) {
	var (
		id    uint64
		login string
	)

	err := row.Scan(&id, &login)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.NoneUser, nil
		}

		return model.NoneUser, err
	}

	return model.NewUser(id, login), nil
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

func (a app) list(page, pageSize uint, sortKey string, sortAsc bool) ([]model.User, uint, error) {
	order := "creation_date DESC"

	if sortKeyMatcher.MatchString(sortKey) {
		order = sortKey

		if !sortAsc {
			order += " DESC"
		}
	}

	offset := (page - 1) * pageSize

	ctx, cancel := context.WithTimeout(context.Background(), sqlTimeout)
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

func (a app) get(id uint64) (model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), sqlTimeout)
	defer cancel()

	row := a.db.QueryRowContext(ctx, getByIDQuery, id)

	return scanUser(row)
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

func (a app) create(o model.User, tx *sql.Tx) (newID uint64, err error) {
	var usedTx *sql.Tx
	if usedTx, err = db.GetTx(a.db, tx); err != nil {
		return
	}

	if usedTx != tx {
		defer func() {
			err = db.EndTx(usedTx, err)
		}()
	}

	ctx, cancel := context.WithTimeout(context.Background(), sqlTimeout)
	defer cancel()

	err = usedTx.QueryRowContext(ctx, insertQuery, o.Login, o.Password).Scan(&newID)
	return
}

const updateQuery = `
UPDATE
  login
SET
  login = $2
WHERE
  id = $1
`

func (a app) update(o model.User, tx *sql.Tx) (err error) {
	var usedTx *sql.Tx
	if usedTx, err = db.GetTx(a.db, tx); err != nil {
		return
	}

	if usedTx != tx {
		defer func() {
			err = db.EndTx(usedTx, err)
		}()
	}

	ctx, cancel := context.WithTimeout(context.Background(), sqlTimeout)
	defer cancel()

	_, err = usedTx.ExecContext(ctx, updateQuery, o.ID, o.Login)

	return
}

const deleteQuery = `
DELETE FROM
  login
WHERE
  id = $1
`

func (a app) delete(o model.User, tx *sql.Tx) (err error) {
	var usedTx *sql.Tx
	if usedTx, err = db.GetTx(a.db, tx); err != nil {
		return
	}

	if usedTx != tx {
		defer func() {
			err = db.EndTx(usedTx, err)
		}()
	}

	ctx, cancel := context.WithTimeout(context.Background(), sqlTimeout)
	defer cancel()

	_, err = usedTx.ExecContext(ctx, deleteQuery, o.ID)
	return
}
