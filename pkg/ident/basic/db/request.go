package db

import (
	"database/sql"

	"github.com/ViBiOh/auth/pkg/model"
)

const readUserQuery = `
SELECT
  id,
  username
FROM
  "user"
WHERE
  username = $1
  AND password = crypt('$2', password)
`

func (a app) getUserFromCredentials(inputUsername, inputPassword string) (model.User, error) {
	var (
		id       uint64
		username string
	)

	if err := a.db.QueryRow(readUserQuery, inputUsername, inputPassword).Scan(&id, &username); err != nil {
		if err == sql.ErrNoRows {
			return model.NoneUser, nil
		}

		return model.NoneUser, err
	}

	return model.NewUser(id, username, ""), nil
}
