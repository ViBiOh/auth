package memory

import (
	"context"
	"strings"

	"github.com/ViBiOh/auth/v2/pkg/ident"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"golang.org/x/crypto/bcrypt"
)

func (a app) Login(_ context.Context, login, password string) (model.User, error) {
	user, ok := a.ident[login]
	if !ok {
		return model.NoneUser, ident.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword(user.password, []byte(password)); err != nil {
		return model.NoneUser, ident.ErrInvalidCredentials
	}

	return user.User, nil
}

func (a app) IsAuthorized(_ context.Context, user model.User, profile string) bool {
	profiles, ok := a.auth[user.ID]
	if !ok {
		return false
	}

	if len(strings.TrimSpace(profile)) == 0 {
		return true
	}

	for _, listedProfile := range profiles {
		if strings.EqualFold(profile, listedProfile) {
			return true
		}
	}

	return false
}
