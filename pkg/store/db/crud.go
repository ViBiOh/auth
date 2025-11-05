package db

import (
	"context"

	"github.com/ViBiOh/auth/v3/pkg/model"
)

func (s Service) DoAtomic(ctx context.Context, action func(context.Context) error) error {
	return s.db.DoAtomic(ctx, action)
}

const insertQuery = `
INSERT INTO
  auth.user
(
  id
)
VALUES (
  $1
)
`

func (s Service) Create(ctx context.Context) (model.User, error) {
	user := model.NewUser("")

	return user, s.db.One(ctx, insertQuery, user.ID)
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
