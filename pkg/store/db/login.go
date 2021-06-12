package db

import (
	"context"
	"database/sql"
	"strings"

	"github.com/ViBiOh/auth/v2/pkg/ident"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
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

func (a app) Login(ctx context.Context, login, password string) (model.User, error) {
	var user model.User

	scanner := func(row *sql.Row) error {
		return row.Scan(&user.ID, &user.Login)
	}

	if err := a.db.Get(ctx, scanner, readUserQuery, strings.ToLower(login), password); err != nil {
		logger.WithField("login", login).Error("unable to login: %s", err.Error())

		if err == sql.ErrNoRows {
			return model.NoneUser, ident.ErrInvalidCredentials
		}
		return model.NoneUser, ident.ErrUnavailableService
	}

	return user, nil
}

const readLoginProfile = `
SELECT
  p.id
FROM
  auth.profile p,
  auth.login_profile lp
WHERE
  p.name = $2
  AND lp.profile_id = p.id
  AND lp.login_id = $1
`

func (a app) IsAuthorized(ctx context.Context, user model.User, profile string) bool {
	var id uint64

	scanner := func(row *sql.Row) error {
		return row.Scan(&id)
	}

	if err := a.db.Get(ctx, scanner, readLoginProfile, user.ID, profile); err != nil {
		logger.WithField("login", user.Login).Error("unable to authorized: %s", err.Error())

		return false
	}

	return true
}
