package basic

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ViBiOh/auth/v3/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/httperror"
)

func (s Service) GetUser(ctx context.Context, w http.ResponseWriter, r *http.Request) (model.User, error) {
	if s.cookie.IsEnabled() {
		claim, err := s.cookie.Get(r, cookieName)
		if err == nil {
			return claim.Content, nil
		}
	}

	login, password, ok := r.BasicAuth()
	if !ok {
		return model.User{}, model.ErrMalformedContent
	}

	user, err := s.provider.GetBasicUser(ctx, login, password)
	if err == nil && s.cookie.IsEnabled() {
		s.cookie.Set(ctx, w, cookieName, user)
	}

	return user, err
}

func (s Service) OnUnauthorized(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, model.ErrMalformedContent) {
		err = nil // We don't want to log it
	}

	w.Header().Add("WWW-Authenticate", fmt.Sprintf("Basic %scharset=\"UTF-8\"", s.realm))
	httperror.Unauthorized(r.Context(), w, err)
}

func (s Service) Logout(w http.ResponseWriter, r *http.Request) {
	s.cookie.Clear(w, cookieName)
}
