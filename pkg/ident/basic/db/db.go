package db

import (
	"database/sql"

	"github.com/ViBiOh/auth/v2/pkg/ident"
	"github.com/ViBiOh/auth/v2/pkg/ident/basic"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
)

var _ basic.UserLogin = App{}

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

const insertUserQuery = `
INSERT INTO login
(
  login,
  password
) VALUES (
  $1,
  crypt($2, gen_salt('bf',8))
)
`

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
