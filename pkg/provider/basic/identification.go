package basic

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/httperror"
)

func (s Service) GetUser(ctx context.Context, r *http.Request) (model.User, error) {
	content := r.Header.Get("Authorization")
	if len(content) == 0 || content[:lenPrefix] != authPrefix {
		return model.User{}, model.ErrMalformedContent
	}

	if len(content) < lenPrefix {
		return model.User{}, model.ErrMalformedContent
	}

	rawData, err := base64.StdEncoding.DecodeString(content[lenPrefix:])
	if err != nil {
		return model.User{}, fmt.Errorf("%s: %w", err, model.ErrMalformedContent)
	}

	data := string(rawData)

	sepIndex := strings.Index(data, ":")
	if sepIndex == -1 {
		return model.User{}, model.ErrMalformedContent
	}

	login := strings.ToLower(data[:sepIndex])
	password := strings.TrimSuffix(data[sepIndex+1:], "\n")

	return s.provider.Login(ctx, r, login, password)
}

func (s Service) OnError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, model.ErrMalformedContent) {
		err = nil // We don't want to log it
	}

	w.Header().Add("WWW-Authenticate", fmt.Sprintf("Basic %scharset=\"UTF-8\"", s.realm))
	httperror.Unauthorized(r.Context(), w, err)
}
