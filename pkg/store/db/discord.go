package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/ViBiOh/auth/v3/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/id"
	"github.com/jackc/pgx/v5"
)

const discordCreateRegistrationQuery = `
INSERT INTO
	auth.discord
(
	id,
	user_id,
	username
) VALUES (
	0,
	$1,
	$2
)
`

func (s Service) CreateDisord(ctx context.Context) (model.User, string, error) {
	user, err := s.Create(ctx)
	if err != nil {
		return user, "", fmt.Errorf("create user: %w", err)
	}

	registrationID := id.New()

	return user, registrationID, s.db.One(ctx, discordCreateRegistrationQuery, user.ID, registrationID)
}

const discordGetUserByIdQuery = `
SELECT
	u.id,
	d.username
FROM
	auth.discord d,
	auth.user u
WHERE
	d.id = $1
	AND d.user_id = u.id
`

const discordGetUserByRegistrationQuery = `
SELECT
	u.id,
	d.username
FROM
	auth.discord d,
	auth.user u
WHERE
	d.username = $1
	AND d.user_id = u.id
`

func (s Service) GetDiscordUser(ctx context.Context, id, registration string) (model.User, error) {
	var item model.User

	query := discordGetUserByIdQuery
	var args any = id

	if len(registration) != 0 {
		query = discordGetUserByRegistrationQuery
		args = registration
	}

	return item, s.db.Get(ctx, func(row pgx.Row) error {
		err := row.Scan(&item.ID, &item.Name)

		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrUnknownUser
		}

		return err
	}, query, args)
}

const discordUpdateUserQuery = `
UPDATE
	auth.discord
SET
	id = $2,
	username = $3,
	avatar = $4
WHERE
	user_id = $1
`

func (s Service) UpdateDiscordUser(ctx context.Context, user model.User, id, username, avatar string) error {
	return s.db.One(ctx, discordUpdateUserQuery, user.ID, id, username, avatar)
}
