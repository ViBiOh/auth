package chooser

import (
	"context"
	"net/http"
	"net/url"

	"github.com/ViBiOh/auth/v3/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
)

type Logout interface {
	Logout(http.ResponseWriter, *http.Request)
}

type Provider struct {
	Auth         model.Authentication
	RegisterPath string
	Kind         model.UserKind
}

type Service struct {
	renderer  *renderer.Service
	providers []Provider
}

func New(renderer *renderer.Service, providers ...Provider) Service {
	return Service{
		providers: providers,
		renderer:  renderer,
	}
}

func (s Service) GetUser(ctx context.Context, w http.ResponseWriter, r *http.Request) (model.User, error) {
	var lastErr error

	for _, p := range s.providers {
		user, err := p.Auth.GetUser(ctx, w, r)
		if err == nil {
			return user, nil
		}
		lastErr = err
	}

	return model.User{}, lastErr
}

func (s Service) OnUnauthorized(w http.ResponseWriter, r *http.Request, err error) {
	redirect := r.URL.String()

	type providerLink struct {
		Name string
		URL  string
	}

	redirection := "?redirect=" + url.QueryEscape(redirect)

	links := make([]providerLink, 0, len(s.providers))
	for _, provider := range s.providers {
		links = append(links, providerLink{
			Name: provider.Kind.String(),
			URL:  provider.RegisterPath + redirection,
		})
	}

	s.renderer.Serve(w, r, renderer.NewPage("auth", http.StatusOK, map[string]any{
		"Providers": links,
	}))
}

func (s Service) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, err := s.GetUser(ctx, w, r)
	if err != nil {
		s.renderer.Redirect(w, r, "/", renderer.NewErrorMessage("unable to logout user: `%s`", err))
		return
	}

	for _, provider := range s.providers {
		if provider.Kind == user.Kind {
			provider.Auth.Logout(w, r)
			return
		}
	}
}
