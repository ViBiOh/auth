package service

import (
	"context"

	"github.com/ViBiOh/auth/v2/pkg/model"
)

type Service struct {
	storage model.Storage
}

func New(storage model.Storage) Service {
	return Service{
		storage: storage,
	}
}

func (s Service) Create(ctx context.Context) (model.User, error) {
	return s.storage.Create(ctx)
}

func (s Service) Delete(ctx context.Context, user model.User) error {
	return s.storage.Delete(ctx, user)
}
