package service

import (
	"context"
	"fmt"

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

func (s Service) Create(ctx context.Context, user model.User) (model.User, error) {
	id, err := s.storage.Create(ctx, user)
	if err != nil {
		return model.User{}, fmt.Errorf("create: %w", err)
	}

	user.ID = id

	return user, nil
}

func (s Service) Delete(ctx context.Context, user model.User) error {
	return s.storage.Delete(ctx, user)
}
