package discord

import (
	"context"
	"net/http"

	"github.com/ViBiOh/auth/v3/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/httperror"
)

func (s Service) IsAuthorized(ctx context.Context, user model.User, profile string) bool {
	return s.provider.IsAuthorized(ctx, user, profile)
}

func (s Service) OnForbidden(w http.ResponseWriter, r *http.Request, user model.User, profile string) {
	if s.onForbidden == nil {
		httperror.Forbidden(r.Context(), w)
	} else {
		s.onForbidden(w, r, user, profile)
	}
}
