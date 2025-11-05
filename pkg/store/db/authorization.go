package db

import (
	"context"
	"log/slog"

	"github.com/ViBiOh/auth/v3/pkg/model"
	"github.com/jackc/pgx/v5"
)

const readLoginProfile = `
SELECT
  p.id
FROM
  auth.profile p,
  auth.user_profile up
WHER
  p.name = $2
  AND up.profile_id = p.id
  AND up.user_id = $1
`

func (s Service) IsAuthorized(ctx context.Context, user model.User, profile string) bool {
	if len(profile) == 0 {
		return true
	}

	var id string

	scanner := func(row pgx.Row) error {
		return row.Scan(&id)
	}

	if err := s.db.Get(ctx, scanner, readLoginProfile, user.ID, profile); err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "unauthorized", slog.String("login", user.Name), slog.Any("error", err))

		return false
	}

	return true
}
