package middleware

import (
	"net/http"

	"github.com/ViBiOh/auth/v3/pkg/model"
	httpmodel "github.com/ViBiOh/httputils/v4/pkg/model"
	"go.opentelemetry.io/otel/trace"
)

var _ httpmodel.Middleware = Service{}.Middleware

type Provider interface{}

type Service struct {
	tracer         trace.Tracer
	identification model.Authentication
	authorization  model.Authorization
}

type ServiceOption func(Service) Service

func WithTracer(tracer trace.Tracer) ServiceOption {
	return func(instance Service) Service {
		instance.tracer = tracer

		return instance
	}
}

func WithAuthorization(authorization model.Authorization) ServiceOption {
	return func(instance Service) Service {
		instance.authorization = authorization

		return instance
	}
}

func New(identification model.Authentication, opts ...ServiceOption) Service {
	service := Service{
		identification: identification,
	}

	for _, option := range opts {
		service = option(service)
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

		user, err := s.identification.GetUser(ctx, r)
		if err != nil {
			s.identification.OnUnauthorized(w, r, err)
			return
		}

		if s.authorization != nil && !s.authorization.IsAuthorized(ctx, r, user) {
			s.authorization.OnForbidden(w, r, user)
			return
		}

		if next != nil {
			next.ServeHTTP(w, r.WithContext(model.StoreUser(ctx, user)))
		}
	})
}
