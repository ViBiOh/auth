package basic

import (
	"database/sql"

	"github.com/ViBiOh/auth/pkg/model"
	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/httputils/pkg/logger"
)

const readUserQuery = `
SELECT
  id,
  username,
  email,
  password
FROM
  "user"
WHERE
  username = $1
`

func (a App) dbLoginUser(login string) *basicUser {
	var (
		id       string
		username string
		email    string
		password string
	)

	if err := a.db.QueryRow(readUserQuery, login).Scan(&id, &username, &email, &password); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}

		logger.Error("%#v", errors.WithStack(err))
		return nil
	}

	return &basicUser{
		model.NewUser(id, username, email, ""),
		[]byte(password),
	}
}
