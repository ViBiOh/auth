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
(
  id
)
VALUES (
  nextval('auth.user_seq')
)
RETURNING id
`

func (s Service) Create(ctx context.Context) (model.User, error) {
	id, err := s.db.Create(ctx, insertQuery)
	return model.User{ID: id}, err
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
