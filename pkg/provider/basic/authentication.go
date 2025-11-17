package basic

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/auth/v3/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/httperror"
)

func (s Service) GetUser(ctx context.Context, r *http.Request) (model.User, error) {
	login, password, err := ExtractCredentials(r)
	if err != nil {
		return model.User{}, err
	}

	return s.provider.GetBasicUser(ctx, login, password)
}

func ExtractCredentials(r *http.Request) (string, string, error) {
	content := r.Header.Get("Authorization")
	if len(content) == 0 || content[:lenPrefix] != authPrefix {
		return "", "", model.ErrMalformedContent
	}

	if len(content) < lenPrefix {
		return "", "", model.ErrMalformedContent
	}

	rawData, err := base64.StdEncoding.DecodeString(content[lenPrefix:])
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", err, model.ErrMalformedContent)
	}

	data := string(rawData)

	sepIndex := strings.Index(data, ":")
	if sepIndex == -1 {
		return "", "", model.ErrMalformedContent
	}

	return strings.ToLower(data[:sepIndex]), strings.TrimSuffix(data[sepIndex+1:], "\n"), nil
}

func (s Service) OnUnauthorized(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, model.ErrMalformedContent) {
		err = nil // We don't want to log it
	}

	w.Header().Add("WWW-Authenticate", fmt.Sprintf("Basic %scharset=\"UTF-8\"", s.realm))
	httperror.Unauthorized(r.Context(), w, err)
}
