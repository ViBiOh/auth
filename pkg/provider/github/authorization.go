package github

import (
	"context"
	"net/http"

	"github.com/ViBiOh/auth/v2/pkg/model"
)

func (s Service) IsAuthorized(_ context.Context, _ model.User, _ string) bool {
	return false
}

func (s Service) OnForbidden(_ http.ResponseWriter, _ *http.Request, _ model.User, _ string) {
}
