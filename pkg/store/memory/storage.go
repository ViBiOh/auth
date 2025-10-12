package memory

import (
	"context"
	"errors"

	"github.com/ViBiOh/auth/v2/pkg/model"
)

func (s Service) Create(_ context.Context, _ model.User) (uint64, error) {
	return 0, errors.New("not updatable")
}

func (s Service) Delete(_ context.Context, _ model.User) error {
	return errors.New("not updatable")
}
