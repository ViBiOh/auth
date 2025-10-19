package memory

import (
	"context"
	"net/http"
	"strings"

	"github.com/ViBiOh/auth/v3/pkg/argon"
	"github.com/ViBiOh/auth/v3/pkg/model"
)

func (s Service) GetBasicUser(_ context.Context, _ *http.Request, login, password string) (model.User, error) {
	user, ok := s.identifications[login]
	if !ok {
		return model.User{}, model.ErrInvalidCredentials
	}

	if strings.HasPrefix(string(user.password), "$argon2id") {
		if argon.CompareHashAndPassword(string(user.password), password) == nil {
			return user.User, nil
		}
	}

	return model.User{}, model.ErrInvalidCredentials
}

func (s Service) IsAuthorized(_ context.Context, user model.User, profile string) bool {
	if len(profile) == 0 {
		return true
	}

	profiles, ok := s.authorizations[user.ID]
	if !ok {
		return false
	}

	for _, listedProfile := range profiles {
		if strings.EqualFold(profile, listedProfile) {
			return true
		}
	}

	return false
}
