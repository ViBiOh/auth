package db

import (
	"database/sql"

	"github.com/ViBiOh/auth/pkg/ident/basic"
	"github.com/ViBiOh/auth/pkg/model"
)

var _ basic.UserLogin = App{}

const readUserQuery = `
SELECT
  id,
  login
FROM
  "user"
WHERE
  login = $1
  AND password = crypt('$2', password)
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
			return model.NoneUser, nil
		}

		return model.NoneUser, err
	}

	return model.NewUser(id, dbLogin), nil
}
