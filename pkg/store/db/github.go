package db

import (
	"context"
	"errors"

	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/id"
	"github.com/jackc/pgx/v5"
)

const githubCreateRegistrationQuery = `
INSERT INTO
  auth.github
(
  user_id,
  login
) VALUES (
  $1,
  $2
)
`

func (s Service) CreateGitHubRegistration(ctx context.Context, user model.User) (string, error) {
	registrationID := id.New()

	return registrationID, s.db.One(ctx, githubCreateRegistrationQuery, user.ID, registrationID)
}

const githubGetUserQuery = `
SELECT
  u.id,
  g.login
FROM
  auth.github g
  auth.user u
WHERE
  g.login = $1
  AND g.user_id = u.id
`

func (s Service) GetGitHubUser(ctx context.Context, registration string) (model.User, error) {
	var item model.User

	return item, s.db.Get(ctx, func(row pgx.Row) error {
		err := row.Scan(&item.ID, &item.Name)

		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrUnknownUser
		}

		return err
	}, githubGetUserQuery, registration)
}

const githubUpdateUserQuery = `
UPDATE
  auth.github
SET
  id = $2,
  login = $3
WHERE
  user_id = $1
`

func (s Service) UpdateGitHubUser(ctx context.Context, user model.User, githubID, githubLogin string) error {
	return s.db.One(ctx, githubUpdateUserQuery, user.ID, githubID, githubLogin)
}
