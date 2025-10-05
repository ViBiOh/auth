package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/ViBiOh/auth/v2/pkg/model"
	httpModel "github.com/ViBiOh/httputils/v4/pkg/model"
)

type Service struct {
	storage model.UpdatableStorage
}

func New(storage model.UpdatableStorage) Service {
	return Service{
		storage: storage,
	}
}

func (s Service) Get(ctx context.Context, ID uint64) (model.User, error) {
	item, err := s.storage.Get(ctx, ID)
	if err != nil {
		return model.User{}, fmt.Errorf("get: %w", err)
	}

	if item.IsZero() {
		return model.User{}, httpModel.WrapNotFound(errors.New("user not found"))
	}

	return item, nil
}

func (s Service) Create(ctx context.Context, user model.User) (model.User, error) {
	id, err := s.storage.Create(ctx, user)
	if err != nil {
		return model.User{}, fmt.Errorf("create: %w", err)
	}

	user.ID = id

	return user, nil
}

func (s Service) Update(ctx context.Context, user model.User) error {
	return s.storage.Update(ctx, user)
}

func (s Service) Delete(ctx context.Context, user model.User) error {
	return s.storage.Delete(ctx, user)
}
