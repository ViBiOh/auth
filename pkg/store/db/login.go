package db

import (
	"context"
	"strings"

	"github.com/ViBiOh/auth/v2/pkg/ident"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/jackc/pgx/v4"
)

const readUserQuery = `
SELECT
  id,
  login
FROM
  auth.login
WHERE
  login = $1
  AND password = crypt($2, password)
`

// Login checks given credentials
func (a App) Login(ctx context.Context, login, password string) (model.User, error) {
	var user model.User

	scanner := func(row pgx.Row) error {
		return row.Scan(&user.ID, &user.Login)
	}

	if err := a.db.Get(ctx, scanner, readUserQuery, strings.ToLower(login), password); err != nil {
		logger.WithField("login", login).Error("login: %s", err.Error())

		if err == pgx.ErrNoRows {
			return model.User{}, ident.ErrInvalidCredentials
		}
		return model.User{}, ident.ErrUnavailableService
	}

	return user, nil
}

const readLoginProfile = `
SELECT
  p.id
FROM
  auth.profile p,
  auth.login_profile lp
WHER
  p.name = $2
  AND lp.profile_id = p.id
  AND lp.login_id = $1
`

// IsAuthorized checks user on profile
func (a App) IsAuthorized(ctx context.Context, user model.User, profile string) bool {
	var id uint64

	scanner := func(row pgx.Row) error {
		return row.Scan(&id)
	}

	if err := a.db.Get(ctx, scanner, readLoginProfile, user.ID, profile); err != nil {
		logger.WithField("login", user.Login).Error("authorized: %s", err.Error())

		return false
	}

	return true
}
