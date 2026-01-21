package oauth

import (
	"context"
	"errors"
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

func (s Service[T, I]) GetUser(ctx context.Context, w http.ResponseWriter, r *http.Request) (model.User, error) {
	claim, err := s.cookie.Get(r, cookieName)
	if err != nil {
		return model.User{}, err
	}

	if len(claim.Content.User.ID) == 0 {
		return model.User{}, errors.New("no content")
	}

	if slices.Contains(updateMethods, r.Method) {
		ctx := r.Context()
		key := updateCacheKey + claim.Content.User.ID

		if content, _ := s.cache.Load(ctx, key); content == nil {
			initialToken := claim.Content.Token.AccessToken

			_, err := s.config.Client(ctx, claim.Content.Token).Get(s.getURL)
			if err != nil {
				return model.User{}, fmt.Errorf("refresh user: %w", err)
			}

			_ = s.cache.Store(ctx, key, time.Now(), updateCheckTTL)

			if initialToken != claim.Content.Token.AccessToken {
				s.cookie.Set(ctx, w, cookieName, model.OAuthClaim{
					Token: claim.Content.Token,
					User:  claim.Content.User,
				})
			}
		}
	}

	return claim.Content.User, nil
}

func (s Service[T, I]) OnUnauthorized(w http.ResponseWriter, r *http.Request, err error) {
	s.redirect(w, r, "", r.URL.String())
}
