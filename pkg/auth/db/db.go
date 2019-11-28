package db

import (
	"database/sql"

	"github.com/ViBiOh/auth/pkg/auth"
	"github.com/ViBiOh/auth/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
)

var _ auth.Provider = App{}

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
