package db

import (
	"context"
	"errors"

	"github.com/ViBiOh/auth/v3/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/id"
	"github.com/jackc/pgx/v5"
)

const createLinkQuery = `
INSERT INTO
  auth.user_link
(
  external_id,
  token,
  description
) VALUES (
  $1,
  $2,
  $3
)
`

func (s Service) CreateLink(ctx context.Context, externalID, description string) (string, error) {
	token := id.New()

	return token, s.db.One(ctx, createLinkQuery, externalID, token, description)
}

const getLinkByTokenQuery = `
SELECT
  external_id,
  description
FROM
  auth.user_link
WHERE
  token = $1
`

func (s Service) GetExternalByToken(ctx context.Context, token string) (model.Link, error) {
	var item model.Link

	return item, s.db.Get(ctx, func(row pgx.Row) error {
		err := row.Scan(&item.ExternalID, &item.Description)

		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrUnknownLink
		}

		return err
	}, getLinkByTokenQuery, token)
}

const updateLinkQuery = `
UPDATE
  auth.user_link
SET
  user_id = $2,
  token = ''
WHERE
  external_id = $1
`

func (s Service) UpdateLink(ctx context.Context, externalID string, user model.User) error {
	return s.db.One(ctx, updateLinkQuery, externalID, user.ID)
}

const deleteLinkQuery = `
DELETE FROM
  auth.user_link
WHERE
  external_id = $1
`

func (s Service) Unlink(ctx context.Context, externalID string) error {
	return s.db.One(ctx, deleteLinkQuery, externalID)
}
