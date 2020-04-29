package db

import (
	"context"
	"database/sql"
	"strings"

	"github.com/ViBiOh/auth/v2/pkg/ident"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/db"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
)

const readUserQuery = `
SELECT
  id,
  login
FROM
  login
WHERE
  login = $1
  AND password = crypt($2, password)
`

func (a app) Login(ctx context.Context, login, password string) (model.User, error) {
	var user model.User

	ctx, cancel := context.WithTimeout(ctx, db.SQLTimeout)
	defer cancel()

	if err := a.db.QueryRowContext(ctx, readUserQuery, strings.ToLower(login), password).Scan(&user.ID, &user.Login); err != nil {
		logger.Error("unable to login %s: %s", login, err.Error())

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
  profile p,
  login_profile lp
WHERE
  p.name = $2
  AND lp.profile_id = p.id
  AND lp.login_id = $1
`

func (a app) IsAuthorized(ctx context.Context, user model.User, profile string) bool {
	var id uint64

	ctx, cancel := context.WithTimeout(ctx, db.SQLTimeout)
	defer cancel()

	if err := a.db.QueryRowContext(ctx, readLoginProfile, user.ID, profile).Scan(&id); err != nil {
		logger.Error("unable to authorized %s: %s", user.Login, err.Error())

		return false
	}

	return true
}
