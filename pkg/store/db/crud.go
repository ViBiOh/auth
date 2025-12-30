package db

import (
	"context"
	"slices"

	"github.com/ViBiOh/auth/v3/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/concurrent"
)

func (s Service) DoAtomic(ctx context.Context, action func(context.Context) error) error {
	return s.db.DoAtomic(ctx, action)
}

const insertQuery = `
INSERT INTO
  auth.user
(
  id
)
VALUES (
  $1
)
`

func (s Service) Create(ctx context.Context, name string) (model.User, error) {
	user := model.NewUser(name)

	return user, s.db.One(ctx, insertQuery, user.ID)
}

func (s Service) List(ctx context.Context, ids ...string) ([]model.User, error) {
	conc := concurrent.NewFailFast(0)

	var discordUsers, githubUsers []model.User

	conc.Go(func() (err error) {
		discordUsers, err = s.listDiscordUsers(ctx, ids...)
		return err
	})

	conc.Go(func() (err error) {
		githubUsers, err = s.listGithubUsers(ctx, ids...)
		return err
	})

	err := conc.Wait()

	return slices.Concat(discordUsers, githubUsers), err
}

const deleteQuery = `
DELETE FROM
  auth.user
WHERE
  id = $1
`

func (s Service) Delete(ctx context.Context, o model.User) error {
	return s.db.One(ctx, deleteQuery, o.ID)
}
