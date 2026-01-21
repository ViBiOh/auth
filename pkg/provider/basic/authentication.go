package basic

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ViBiOh/auth/v3/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/httperror"
)

func (s Service) GetUser(ctx context.Context, _ http.ResponseWriter, r *http.Request) (model.User, error) {
	login, password, ok := r.BasicAuth()
	if !ok {
		return model.User{}, model.ErrMalformedContent
	}

	return s.provider.GetBasicUser(ctx, login, password)
}

func (s Service) OnUnauthorized(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, model.ErrMalformedContent) {
		err = nil // We don't want to log it
	}

	w.Header().Add("WWW-Authenticate", fmt.Sprintf("Basic %scharset=\"UTF-8\"", s.realm))
	httperror.Unauthorized(r.Context(), w, err)
}
