package discord

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/ViBiOh/auth/v3/pkg/model"
)

const updateCheckTTL = time.Hour * 2

var updateMethods = []string{
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
}

func (s Service) GetUser(ctx context.Context, r *http.Request) (model.User, error) {
	claim, err := s.cookie.Get(r, cookieName)
	if err != nil {
		return model.User{}, err
	}

	if slices.Contains(updateMethods, r.Method) {
		ctx := r.Context()
		key := updateCacheKey + claim.User.ID

		if content, _ := s.cache.Load(ctx, key); content == nil {
			if _, err := s.config.Client(ctx, claim.Token).Get("https://discord.com/api/users/@me"); err != nil {
				return model.User{}, fmt.Errorf("refresh user: %w", err)
			}

			_ = s.cache.Store(ctx, key, time.Now(), updateCheckTTL)
		}
	}

	return claim.User, nil
}

func (s Service) OnUnauthorized(w http.ResponseWriter, r *http.Request, err error) {
	s.redirect(w, r, "", r.URL.String())
}
