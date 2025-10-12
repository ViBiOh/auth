package db

import (
	"context"
	"errors"
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
  auth.user
WHERE
  id = $1
`

func (s Service) Get(ctx context.Context, id uint64) (model.User, error) {
	var item model.User

	scanner := func(row pgx.Row) error {
		err := row.Scan(&item.ID, &item.Login)

		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrUnknownUser
		}

		return err
	}

	return item, s.db.Get(ctx, scanner, getByIDQuery, id)
}

const insertQuery = `
INSERT INTO
  auth.user
(
  login
) VALUES (
  $1
) RETURNING id
`

func (s Service) Create(ctx context.Context, o model.User) (uint64, error) {
	return s.db.Create(ctx, insertQuery, strings.ToLower(o.Login))
}

const updateQuery = `
UPDATE
  auth.user
SET
  login = $2
WHERE
  id = $1
`

func (s Service) Update(ctx context.Context, o model.User) error {
	return s.db.One(ctx, updateQuery, o.ID, strings.ToLower(o.Login))
}

const deleteQuery = `
DELETE FROM
  auth.user
WHERE
  id = $1
`

func (s Service) Delete(ctx context.Context, o model.User) error {
	return s.db.One(ctx, deleteQuery, o.ID)
}
