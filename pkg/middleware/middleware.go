package middleware

import (
	"net/http"

	"github.com/ViBiOh/auth/v3/pkg/model"
	httpmodel "github.com/ViBiOh/httputils/v4/pkg/model"
	"go.opentelemetry.io/otel/trace"
)

var _ httpmodel.Middleware = Service{}.Middleware

type Provider interface {
	model.Identification
	model.Authorization
}

type Service struct {
	tracer   trace.Tracer
	provider Provider
	profile  string
}

func New(provider Provider, profile string, tracerProvider trace.TracerProvider) Service {
	service := Service{
		provider: provider,
		profile:  profile,
	}

	if tracerProvider != nil {
		service.tracer = tracerProvider.Tracer("auth")
	}

	return service
}

func (s Service) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		ctx := r.Context()

		user, err := s.provider.GetUser(ctx, r)
		if err != nil {
			s.provider.OnError(w, r, err)
			return
		}

		if !s.provider.IsAuthorized(ctx, user, s.profile) {
			s.provider.OnForbidden(w, r, user, s.profile)
			return
		}

		if next != nil {
			next.ServeHTTP(w, r.WithContext(model.StoreUser(ctx, user)))
		}
	})
}
