package db

import (
	"database/sql"

	"github.com/ViBiOh/auth/v2/pkg/auth"
	"github.com/ViBiOh/auth/v2/pkg/ident"
	"github.com/ViBiOh/auth/v2/pkg/ident/basic"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
)

var (
	_ auth.Provider   = App{}
	_ basic.UserLogin = App{}
)

// App of package
type App struct {
	db *sql.DB
}

// New creates new App from dependencies
func New(db *sql.DB) App {
	return App{
		db: db,
	}
}

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

// Login user with its credentials
func (a App) Login(login, password string) (model.User, error) {
	var (
		id      uint64
		dbLogin string
	)

	if err := a.db.QueryRow(readUserQuery, login, password).Scan(&id, &dbLogin); err != nil {
		if err == sql.ErrNoRows {
			return model.NoneUser, ident.ErrInvalidCredentials
		}

		logger.Error("%s", err.Error())
		return model.NoneUser, ident.ErrUnavailableService
	}

	return model.NewUser(id, dbLogin), nil
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

// IsAuthorized checks if User is authorized
func (a App) IsAuthorized(user model.User, profile string) bool {
	var id uint64

	if err := a.db.QueryRow(readLoginProfile, user.ID, profile).Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return false
		}

		logger.Error("%s", err.Error())
		return false
	}

	return true
}
