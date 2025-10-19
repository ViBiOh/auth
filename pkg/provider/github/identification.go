package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/ViBiOh/auth/v3/pkg/model"
	"github.com/golang-jwt/jwt/v5"
)

const updateCheckTTL = time.Hour * 2

var updateMethods = []string{
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
}

func (s Service) GetUser(ctx context.Context, r *http.Request) (model.User, error) {
	auth, err := r.Cookie(cookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return model.User{}, model.ErrMalformedContent
		}

		return model.User{}, fmt.Errorf("get auth cookie: %w", err)
	}

	var claim AuthClaims

	if _, err = jwt.ParseWithClaims(auth.Value, &claim, s.jwtKeyFunc, jwt.WithValidMethods(signValidMethods)); err != nil {
		return model.User{}, fmt.Errorf("parse JWT: %w", err)
	}

	if slices.Contains(updateMethods, r.Method) {
		ctx := r.Context()
		key := updateCacheKey + strconv.FormatUint(claim.User.ID, 10)

		if content, _ := s.cache.Load(ctx, key); content == nil {
			if _, err := s.config.Client(ctx, claim.Token).Get("https://api.github.com/user"); err != nil {
				return model.User{}, fmt.Errorf("refresh user: %w", err)
			}

			_ = s.cache.Store(ctx, key, time.Now(), updateCheckTTL)
		}
	}

	return claim.User, nil
}

func (s Service) OnError(w http.ResponseWriter, r *http.Request, err error) {
	s.redirect(w, r, "")
}
