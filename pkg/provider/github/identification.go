package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/golang-jwt/jwt/v5"
)

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

	return claim.User, nil
}

func (s Service) OnError(w http.ResponseWriter, r *http.Request, err error) {
	s.redirect(w, r, "")
}
