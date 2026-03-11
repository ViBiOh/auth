package chooser

import (
	"context"
	"net/http"
	"net/url"

	"github.com/ViBiOh/auth/v3/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
)

type Provider struct {
	Auth         model.Authentication
	Name         string
	RegisterPath string
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

	links := make([]providerLink, len(s.providers))
	for i, p := range s.providers {
		links[i] = providerLink{
			Name: p.Name,
			URL:  p.RegisterPath + "?redirect=" + url.QueryEscape(redirect),
		}
	}

	s.renderer.Serve(w, r, renderer.NewPage("auth", http.StatusOK, map[string]any{
		"Providers": links,
	}))
}
