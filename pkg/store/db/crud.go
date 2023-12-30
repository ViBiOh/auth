package db

import (
	"context"
	"strings"

	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/jackc/pgx/v5"
)

func (s Service) DoAtomic(ctx context.Context, action func(context.Context) error) error {
	return s.db.DoAtomic(ctx, action)
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

func (s Service) Get(ctx context.Context, id uint64) (model.User, error) {
	var item model.User
	scanner := func(row pgx.Row) error {
		err := row.Scan(&item.ID, &item.Login)
		if err == pgx.ErrNoRows {
			item = model.User{}
			return nil
		}

		return err
	}

	return item, s.db.Get(ctx, scanner, getByIDQuery, id)
}

const insertQuery = `
INSERT INTO
  auth.login
(
  login,
  password
) VALUES (
  $1,
  crypt($2, gen_salt('bf',10))
) RETURNING id
`

func (s Service) Create(ctx context.Context, o model.User) (uint64, error) {
	return s.db.Create(ctx, insertQuery, strings.ToLower(o.Login), o.Password)
}

const updateQuery = `
UPDATE
  auth.login
SET
  login = $2
WHERE
  id = $1
`

func (s Service) Update(ctx context.Context, o model.User) error {
	return s.db.One(ctx, updateQuery, o.ID, strings.ToLower(o.Login))
}

const updatePasswordQuery = `
UPDATE
  auth.login
SET
  password = crypt($2, gen_salt('bf',10))
WHERE
  id = $1
`

func (s Service) UpdatePassword(ctx context.Context, o model.User) error {
	return s.db.One(ctx, updatePasswordQuery, o.ID, o.Password)
}

const deleteQuery = `
DELETE FROM
  auth.login
WHERE
  id = $1
`

func (s Service) Delete(ctx context.Context, o model.User) error {
	return s.db.One(ctx, deleteQuery, o.ID)
}
