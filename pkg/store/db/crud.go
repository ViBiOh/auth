package db

import (
	"context"
	"strings"

	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/jackc/pgx/v5"
)

// DoAtomic do things in a transaction
func (a Service) DoAtomic(ctx context.Context, action func(context.Context) error) error {
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

// Get a user
func (a Service) Get(ctx context.Context, id uint64) (model.User, error) {
	var item model.User
	scanner := func(row pgx.Row) error {
		err := row.Scan(&item.ID, &item.Login)
		if err == pgx.ErrNoRows {
			item = model.User{}
			return nil
		}

		return err
	}

	return item, a.db.Get(ctx, scanner, getByIDQuery, id)
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

// Create a user
func (a Service) Create(ctx context.Context, o model.User) (uint64, error) {
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
func (a Service) Update(ctx context.Context, o model.User) error {
	return a.db.One(ctx, updateQuery, o.ID, strings.ToLower(o.Login))
}

const updatePasswordQuery = `
UPDATE
  auth.login
SET
  password = crypt($2, gen_salt('bf',8))
WHERE
  id = $1
`

// UpdatePassword of a user
func (a Service) UpdatePassword(ctx context.Context, o model.User) error {
	return a.db.One(ctx, updatePasswordQuery, o.ID, o.Password)
}

const deleteQuery = `
DELETE FROM
  auth.login
WHERE
  id = $1
`

// Delete an user
func (a Service) Delete(ctx context.Context, o model.User) error {
	return a.db.One(ctx, deleteQuery, o.ID)
}
