package service

import (
	"context"
	"net/http"
	"strings"

	"github.com/ViBiOh/auth/pkg/model"
	"github.com/ViBiOh/auth/pkg/provider"
	"github.com/ViBiOh/httputils/pkg/httperror"
)

// GetUser get user from given auth content
func (a App) GetUser(ctx context.Context, authContent string) (*model.User, error) {
	if authContent == `` {
		return nil, provider.ErrEmptyAuthorization
	}

	parts := strings.SplitN(authContent, ` `, 2)
	if len(parts) != 2 {
		return nil, provider.ErrMalformedAuth
	}

	for _, p := range a.providers {
		if parts[0] == p.GetName() {
			user, err := p.GetUser(ctx, parts[1])
			if err != nil {
				return nil, err
			}

			return user, nil
		}
	}

	return nil, provider.ErrUnknownAuthType
}

// OnError handle error for service app
func (a App) OnError(w http.ResponseWriter, r *http.Request, err error) {
	if len(a.providers) > 0 {
		a.providers[0].OnLoginError(w, r, err)
		return
	}

	httperror.Unauthorized(w, err)
}
