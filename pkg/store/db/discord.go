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
  username,
  avatar
) VALUES (
  $1,
  $2,
  $3,
  ''
)
`

func (s Service) CreateDiscord(ctx context.Context, username string) (model.User, string, error) {
	user, err := s.Create(ctx)
	if err != nil {
		return user, "", fmt.Errorf("create user: %w", err)
	}

	registrationID := id.New()

	return user, registrationID, s.db.One(ctx, discordCreateRegistrationQuery, registrationID, user.ID, username)
}

const discordGetUserByIdQuery = `
SELECT
  username,
  id,
  avatar,
  json_agg(user_id) as user_ids
FROM
  auth.discord
WHERE
  id = $1
GROUP BY
  username, id, avatar
`

func (s Service) GetDiscordUser(ctx context.Context, id string) (model.User, error) {
	var item model.User

	return item, s.db.Get(ctx, func(row pgx.Row) error {
		var discordID, avatar string
		err := row.Scan(&item.Name, &discordID, &avatar, &item.Aliases)

		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrUnknownUser
		}

		item.ID = item.Aliases[0]
		item.Image = getDiscordImageURL(discordID, avatar)

		return err
	}, discordGetUserByIdQuery, id)
}

const discordListUsers = `
SELECT
  user_id,
  username,
  id,
  avatar
FROM
  auth.discord
WHERE
  user_id = ANY($1)
`

func (s Service) listDiscordUsers(ctx context.Context, userIDs ...string) ([]model.User, error) {
	var items []model.User

	return items, s.db.List(ctx, func(rows pgx.Rows) error {
		var discordID, avatar string
		var item model.User

		if err := rows.Scan(&item.ID, &item.Name, &discordID, &avatar); err != nil {
			return fmt.Errorf("scan: %w", err)
		}

		item.Image = getDiscordImageURL(discordID, avatar)
		items = append(items, item)

		return nil
	}, discordListUsers, userIDs)
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

func (s Service) UpdateDiscordUser(ctx context.Context, user model.User, id, username, avatar string) (model.User, error) {
	user.Name = username
	user.Image = getDiscordImageURL(id, avatar)

	if err := s.db.One(ctx, discordUpdateUserQuery, user.ID, id, username, avatar); err != nil {
		return user, fmt.Errorf("update: %w", err)
	}

	return s.GetDiscordUser(ctx, id)
}

func getDiscordImageURL(id, avatar string) string {
	if len(id) == 0 || len(avatar) == 0 {
		return ""
	}

	return fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.webp", id, avatar)
}
