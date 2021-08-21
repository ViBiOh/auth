package db

import (
	"context"
	"database/sql"
	"strings"

	"github.com/ViBiOh/auth/v2/pkg/model"
)

// DoAtomic do things in a transaction
func (a App) DoAtomic(ctx context.Context, action func(context.Context) error) error {
	return a.db.DoAtomic(ctx, action)
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

// Get get an user
func (a App) Get(ctx context.Context, id uint64) (model.User, error) {
	var item model.User
	scanner := func(row *sql.Row) error {
		err := row.Scan(&item.ID, &item.Login)
		if err == sql.ErrNoRows {
			item = model.User{}
			return nil
		}

		return err
	}

	err := a.db.Get(ctx, scanner, getByIDQuery, id)
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

// Create an user
func (a App) Create(ctx context.Context, o model.User) (uint64, error) {
	return a.db.Create(ctx, insertQuery, strings.ToLower(o.Login), o.Password)
}

const updateQuery = `
UPDATE
  auth.login
SET
  login = $2
WHERE
  id = $1
`

// Update user
func (a App) Update(ctx context.Context, o model.User) error {
	return a.db.Exec(ctx, updateQuery, o.ID, strings.ToLower(o.Login))
}

const updatePasswordQuery = `
UPDATE
  auth.login
SET
  password = crypt($2, gen_salt('bf',8))
WHERE
  id = $1
`

// UpdatePassword of an user
func (a App) UpdatePassword(ctx context.Context, o model.User) error {
	return a.db.Exec(ctx, updatePasswordQuery, o.ID, o.Password)
}

const deleteQuery = `
DELETE FROM
  auth.login
WHERE
  id = $1
`

// Delete an user
func (a App) Delete(ctx context.Context, o model.User) error {
	return a.db.Exec(ctx, deleteQuery, o.ID)
}
