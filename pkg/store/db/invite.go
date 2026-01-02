package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/ViBiOh/auth/v3/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/id"
	"github.com/jackc/pgx/v5"
)

const createInviteQuery = `
INSERT INTO
  auth.invite
(
  user_id,
  token,
  description
) VALUES (
  $1,
  $2,
  $3
)
`

func (s Service) CreateInvite(ctx context.Context, description string) (model.User, string, error) {
	token := id.New()

	user, err := s.Create(ctx, description)
	if err != nil {
		return user, token, fmt.Errorf("create user: %w", err)
	}

	return user, token, s.db.One(ctx, createInviteQuery, user.ID, token, description)
}

const getInviteByTokenQuery = `
SELECT
  user_id,
  description
FROM
  auth.invite
WHERE
  token = $1
`

func (s Service) GetInviteByToken(ctx context.Context, token string) (model.User, error) {
	var item model.User

	return item, s.db.Get(ctx, func(row pgx.Row) error {
		err := row.Scan(&item.ID, &item.Name)

		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrUnknownUser
		}

		return err
	}, getInviteByTokenQuery, token)
}

const listInviteQuery = `
SELECT
  user_id,
  description
FROM
  auth.invite
WHERE
  user_id = ANY($1)
`

func (s Service) listInviteUsers(ctx context.Context, userIDs ...string) ([]model.User, error) {
	var items []model.User

	return items, s.db.List(ctx, func(rows pgx.Rows) error {
		var item model.User

		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			return fmt.Errorf("scan: %w", err)
		}

		items = append(items, item)

		return nil
	}, listInviteQuery, userIDs)
}

const deleteInviteQuery = `
DELETE FROM
  auth.invite
WHERE
  user_id = $1
`

func (s Service) DeleteInvite(ctx context.Context, o model.User) error {
	return s.db.One(ctx, deleteInviteQuery, o.ID)
}
