package memory

import (
	"context"
	"strings"

	"github.com/ViBiOh/auth/v2/pkg/argon"
	"github.com/ViBiOh/auth/v2/pkg/ident"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"golang.org/x/crypto/bcrypt"
)

func (s Service) Login(_ context.Context, login, password string) (model.User, error) {
	user, ok := s.ident[login]
	if !ok {
		return model.User{}, ident.ErrInvalidCredentials
	}

	switch {
	case strings.HasPrefix(string(user.password), "$argon2id"):
		if argon.CompareHashAndPassword(string(user.password), password) == nil {
			return user.User, nil
		}

	default:
		rawPassword := []byte(password)
		if len(rawPassword) > ident.MaxPasswordLength {
			return model.User{}, ident.ErrTooLongPassword
		}

		if bcrypt.CompareHashAndPassword(user.password, rawPassword) == nil {
			return user.User, nil
		}
	}

	return model.User{}, ident.ErrInvalidCredentials
}

func (s Service) IsAuthorized(_ context.Context, user model.User, profile string) bool {
	profiles, ok := s.auth[user.ID]
	if !ok {
		return false
	}

	if len(profile) == 0 {
		return true
	}

	for _, listedProfile := range profiles {
		if strings.EqualFold(profile, listedProfile) {
			return true
		}
	}

	return false
}
