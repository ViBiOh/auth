package db

import (
	"context"

	"github.com/ViBiOh/auth/v2/pkg/model"
)

func (s Service) DoAtomic(ctx context.Context, action func(context.Context) error) error {
	return s.db.DoAtomic(ctx, action)
}

const insertQuery = `
INSERT INTO
  auth.user
RETURNING id
`

func (s Service) Create(ctx context.Context, o model.User) (uint64, error) {
	return s.db.Create(ctx, insertQuery)
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
