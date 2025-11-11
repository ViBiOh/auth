package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/ViBiOh/auth/v3/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/id"
	"github.com/jackc/pgx/v5"
)

const githubCreateRegistrationQuery = `
INSERT INTO
  auth.github
(
  id,
  user_id,
  login
) VALUES (
  0,
  $1,
  $2
)
`

func (s Service) CreateGithub(ctx context.Context) (model.User, string, error) {
	user, err := s.Create(ctx)
	if err != nil {
		return user, "", fmt.Errorf("create user: %w", err)
	}

	registrationID := id.New()

	return user, registrationID, s.db.One(ctx, githubCreateRegistrationQuery, user.ID, registrationID)
}

const githubGetUserByIdQuery = `
SELECT
  u.id,
  g.login
FROM
  auth.github g,
  auth.user u
WHERE
  g.id = $1
  AND g.user_id = u.id
`

const githubGetUserByRegistrationQuery = `
SELECT
  u.id,
  g.login,
  g.id
FROM
  auth.github g,
  auth.user u
WHERE
  g.login = $1
  AND g.user_id = u.id
`

func (s Service) GetGitHubUser(ctx context.Context, id uint64, registration string) (model.User, error) {
	var item model.User

	query := githubGetUserByIdQuery
	var args any = id

	if len(registration) != 0 {
		query = githubGetUserByRegistrationQuery
		args = registration
	}

	return item, s.db.Get(ctx, func(row pgx.Row) error {
		var githubID string

		err := row.Scan(&item.ID, &item.Name, &githubID)

		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrUnknownUser
		}

		item.Image = getGitHubImageURL(githubID)

		return err
	}, query, args)
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

func (s Service) UpdateGitHubUser(ctx context.Context, user model.User, githubID, githubLogin string) (model.User, error) {
	user.Name = githubLogin
	user.Image = getGitHubImageURL(githubID)

	return user, s.db.One(ctx, githubUpdateUserQuery, user.ID, githubID, githubLogin)
}

func getGitHubImageURL(id string) string {
	return "https://avatars.githubusercontent.com/u/" + id
}
