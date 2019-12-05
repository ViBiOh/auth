package service

import (
	"database/sql"
	"fmt"

	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/db"
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
			return model.NoneUser, err
		}

		return model.NoneUser, err
	}

	return model.NewUser(id, login), nil
}

func scanUsers(rows *sql.Rows) ([]model.User, uint, error) {
	var (
		id         uint64
		login      string
		totalCount uint
	)

	list := make([]model.User, 0)

	for rows.Next() {
		if err := rows.Scan(&id, &login, &totalCount); err != nil {
			if err == sql.ErrNoRows {
				return nil, 0, err
			}

			return nil, 0, err
		}

		list = append(list, model.NewUser(id, login))
	}

	return list, totalCount, nil
}

const listQuery = `
SELECT
  id,
  login,
  count(*) OVER() AS full_count
FROM
  profile
WHERE
  id = $1
ORDER BY $4
LIMIT $2
OFFSET $3
`

func (a app) list(page, pageSize uint, sortKey string, sortAsc bool) ([]model.User, uint, error) {
	order := "creation_date DESC"

	if sortKey != "" {
		order = sortKey
	}

	if !sortAsc {
		order = fmt.Sprintf("%s DESC", order)
	}

	offset := (page - 1) * pageSize

	rows, err := a.db.Query(listQuery, pageSize, offset, order)
	if err != nil {
		return nil, 0, err
	}

	defer func() {
		err = db.RowsClose(rows, err)
	}()

	list, totalCount, err := scanUsers(rows)
	if err != nil {
		return nil, 0, err
	}

	return list, totalCount, nil
}

const getByIDQuery = `
SELECT
  id,
  login
FROM
  profile
WHERE
  id = $1
`

func (a app) get(id uint64) (model.User, error) {
	row := a.db.QueryRow(getByIDQuery, id)

	return scanUser(row)
}

const insertQuery = `
INSERT INTO
  profile
(
  login,
) VALUES (
  $1
)
`

func (a app) create(o model.User, tx *sql.Tx) (id uint64, err error) {
	var usedTx *sql.Tx
	if usedTx, err = db.GetTx(a.db, tx); err != nil {
		return
	}

	if usedTx != tx {
		defer func() {
			err = db.EndTx(usedTx, err)
		}()
	}

	result, insertErr := usedTx.Exec(insertQuery, o.Login)
	if insertErr != nil {
		err = insertErr
		return
	}

	newID, idErr := result.LastInsertId()
	if idErr != nil {
		err = idErr
		return
	}

	id = uint64(newID)

	return
}

const updateQuery = `
UPDATE
  profile
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

	_, err = usedTx.Exec(updateQuery, o.ID, o.Login)

	return
}

const deleteQuery = `
DELETE FROM
  profile
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

	_, err = usedTx.Exec(deleteQuery, o.ID, o.ID)
	return
}
