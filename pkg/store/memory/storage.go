package memory

import (
	"context"
	"errors"

	"github.com/ViBiOh/auth/v3/pkg/model"
)

func (s Service) Create(_ context.Context) (model.User, error) {
	return model.User{}, errors.New("not updatable")
}

func (s Service) Delete(_ context.Context, _ model.User) error {
	return errors.New("not updatable")
}
