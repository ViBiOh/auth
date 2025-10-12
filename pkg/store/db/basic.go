package db

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/ViBiOh/auth/v2/pkg/argon"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

const basicUserQuery = `
SELECT
  u.id,
  u.login,
  b.password
FROM
  auth.user u,
  auth.basic b
WHERE
  u.login = $1
  AND u.id = b.user_id
`

func (s Service) GetBasicUser(ctx context.Context, _ *http.Request, login, password string) (model.User, error) {
	var user model.User
	var userPassword string

	scanner := func(row pgx.Row) error {
		return row.Scan(&user.ID, &user.Name, &userPassword)
	}

	if err := s.db.Get(ctx, scanner, basicUserQuery, strings.ToLower(login)); err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "login", slog.String("login", login), slog.Any("error", err))

		if err == pgx.ErrNoRows {
			return model.User{}, model.ErrInvalidCredentials
		}

		return model.User{}, model.ErrUnavailableService
	}

	switch {
	case strings.HasPrefix(userPassword, "$argon2id"):
		if argon.CompareHashAndPassword(userPassword, password) == nil {
			return user, nil
		}

	default:
		if bcrypt.CompareHashAndPassword([]byte(userPassword), []byte(password)) == nil {
			if err := s.UpdatePassword(ctx, user, password); err != nil {
				slog.LogAttrs(ctx, slog.LevelError, "update password to argon2", slog.Any("error", err))
			}

			return user, nil
		}
	}

	return model.User{}, model.ErrInvalidCredentials
}

const insertPasswordQuery = `
INSERT INTO
  auth.basic
(
  user_id,
  password
) VALUES (
  $1,
  $2
)
`

func (s Service) SavePassword(ctx context.Context, user model.User, password string) error {
	password, err := argon.GenerateFromPassword(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	return s.db.One(ctx, insertPasswordQuery, user.ID, password)
}

const updatePasswordQuery = `
UPDATE
  auth.basic
SET
  password = $2
WHERE
  user_id = $1
`

func (s Service) UpdatePassword(ctx context.Context, user model.User, password string) error {
	password, err := argon.GenerateFromPassword(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	return s.db.One(ctx, updatePasswordQuery, user.ID, password)
}
