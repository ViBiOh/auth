package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/ViBiOh/auth/v2/pkg/auth"
	"github.com/ViBiOh/auth/v2/pkg/ident"
	"github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/httperror"
	httpmodel "github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
	"go.opentelemetry.io/otel/trace"
)

var (
	_ httpmodel.Middleware = Service{}.Middleware

	ErrEmptyAuth          = errors.New("empty authorization content")
	ErrNoMatchingProvider = errors.New("no matching provider for Authorization content")
)

type Service struct {
	tracer         trace.Tracer
	authProvider   auth.Provider
	identProviders []ident.Provider
}

func New(authProvider auth.Provider, tracerProvider trace.TracerProvider, identProviders ...ident.Provider) Service {
	service := Service{
		authProvider:   authProvider,
		identProviders: identProviders,
	}

	if tracerProvider != nil {
		service.tracer = tracerProvider.Tracer("auth")
	}

	return service
}

func (s Service) Middleware(next http.Handler) http.Handler {
	if len(s.identProviders) == 0 {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		provider, user, err := s.IsAuthenticated(r)
		if err != nil {
			onHandlerFail(w, r, err, provider)
			return
		}

		if next != nil {
			next.ServeHTTP(w, r.WithContext(model.StoreUser(r.Context(), user)))
		}
	})
}

func (s Service) IsAuthenticated(r *http.Request) (ident.Provider, model.User, error) {
	if len(s.identProviders) == 0 {
		return nil, model.User{}, ErrNoMatchingProvider
	}

	var err error

	ctx, end := telemetry.StartSpan(r.Context(), s.tracer, "check_auth", trace.WithSpanKind(trace.SpanKindInternal))
	defer end(&err)

	authContent := strings.TrimSpace(r.Header.Get("Authorization"))
	if len(authContent) == 0 {
		return s.identProviders[0], model.User{}, ErrEmptyAuth
	}

	for _, provider := range s.identProviders {
		if !provider.IsMatching(authContent) {
			continue
		}

		user, err := provider.GetUser(ctx, authContent)
		return provider, user, err
	}

	return nil, model.User{}, ErrNoMatchingProvider
}

func (s Service) IsAuthorized(ctx context.Context, profile string) bool {
	if s.authProvider == nil {
		return false
	}

	return s.authProvider.IsAuthorized(ctx, model.ReadUser(ctx), profile)
}

func onHandlerFail(w http.ResponseWriter, r *http.Request, err error, provider ident.Provider) {
	if err == auth.ErrForbidden {
		httperror.Forbidden(r.Context(), w)
	} else if provider != nil {
		provider.OnError(w, r, err)
	} else {
		httperror.Unauthorized(r.Context(), w, err)
	}
}
