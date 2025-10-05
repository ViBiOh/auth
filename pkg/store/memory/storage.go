package memory

import (
	"context"

	"github.com/ViBiOh/auth/v2/pkg/model"
)

func (s Service) Get(ctx context.Context, id uint64) (model.User, error) {
	for _, user := range s.identifications {
		if user.ID == id {
			return user.User, nil
		}
	}

	return model.User{}, model.ErrInvalidCredentials
}
