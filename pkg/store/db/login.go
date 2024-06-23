package db

import (
	"context"
	"log/slog"
	"strings"

	"github.com/ViBiOh/auth/v2/pkg/argon"
	"github.com/ViBiOh/auth/v2/pkg/ident"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

const readUserQuery = `
SELECT
  id,
  login,
  password
FROM
  auth.login
WHERE
  login = $1
`

func (s Service) Login(ctx context.Context, login, password string) (model.User, error) {
	var user model.User

	scanner := func(row pgx.Row) error {
		return row.Scan(&user.ID, &user.Login, &user.Password)
	}

	if err := s.db.Get(ctx, scanner, readUserQuery, strings.ToLower(login)); err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "login", slog.String("login", login), slog.Any("error", err))

		if err == pgx.ErrNoRows {
			return model.User{}, ident.ErrInvalidCredentials
		}
		return model.User{}, ident.ErrUnavailableService
	}

	switch {
	case strings.HasPrefix(string(user.Password), "$argon2id"):
		if argon.CompareHashAndPassword(user.Password, password) == nil {
			user.Password = ""

			return user, nil
		}

	default:
		if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) == nil {
			user.Password = password

			if err := s.UpdatePassword(ctx, user); err != nil {
				slog.LogAttrs(ctx, slog.LevelError, "update password to argon2", slog.Any("error", err))
			}

			user.Password = ""

			return user, nil
		}
	}

	return model.User{}, ident.ErrInvalidCredentials
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

func (s Service) IsAuthorized(ctx context.Context, user model.User, profile string) bool {
	var id uint64

	scanner := func(row pgx.Row) error {
		return row.Scan(&id)
	}

	if err := s.db.Get(ctx, scanner, readLoginProfile, user.ID, profile); err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "authorized", slog.String("login", user.Login), slog.Any("error", err))

		return false
	}

	return true
}
