package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/ViBiOh/auth/v3/pkg/model"
	"github.com/jackc/pgx/v5"
)

const googleCreateRegistrationQuery = `
INSERT INTO
  auth.google
(
  id,
  user_id,
  name,
  picture
) VALUES (
  $1,
  $2,
  $3,
  $4
)
`

func (s Service) CreateGoogle(ctx context.Context, invite model.User, user model.GoogleUser) (model.User, error) {
	invite.Name = user.Name
	invite.Kind = model.Google
	invite.Image = user.Picture

	return invite, s.db.One(ctx, googleCreateRegistrationQuery, user.Sub, invite.ID, user.Name, user.Picture)
}

const googleGetUserByIdQuery = `
SELECT
  user_id,
  name,
  picture
FROM
  auth.google
WHERE
  id = $1
`

func (s Service) GetGoogleUser(ctx context.Context, id string) (model.User, error) {
	var item model.User

	return item, s.db.Get(ctx, func(row pgx.Row) error {
		err := row.Scan(&item.ID, &item.Name, &item.Image)

		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrUnknownUser
		}

		item.Kind = model.Google

		return err
	}, googleGetUserByIdQuery, id)
}

const googleListUsers = `
SELECT
  user_id,
  name,
  picture
FROM
  auth.google
WHERE
  user_id = ANY($1)
`

func (s Service) listGoogleUsers(ctx context.Context, userIDs ...string) ([]model.User, error) {
	var items []model.User

	return items, s.db.List(ctx, func(rows pgx.Rows) error {
		var item model.User

		if err := rows.Scan(&item.ID, &item.Name, &item.Image); err != nil {
			return fmt.Errorf("scan: %w", err)
		}

		item.Kind = model.Google
		items = append(items, item)

		return nil
	}, googleListUsers, userIDs)
}

const googleUpdateUserQuery = `
UPDATE
  auth.google
SET
  id = $2,
  name = $3,
  picture = $4
WHERE
  user_id = $1
`

func (s Service) UpdateGoogleUser(ctx context.Context, user model.User, googleID, name, picture string) (model.User, error) {
	user.Name = name
	user.Image = picture

	if err := s.db.One(ctx, googleUpdateUserQuery, user.ID, googleID, name, picture); err != nil {
		return user, fmt.Errorf("update: %w", err)
	}

	return s.GetGoogleUser(ctx, googleID)
}
