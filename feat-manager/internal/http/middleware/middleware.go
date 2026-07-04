package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/shivam/featfz/feat-manager/internal/http/requestctx"
	"github.com/shivam/featfz/feat-manager/internal/http/response"
	"github.com/shivam/featfz/feat-manager/internal/service"
)

const RequestIDHeader = "X-Request-ID"

type Middleware func(http.Handler) http.Handler

func Chain(next http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		next = middlewares[i](next)
	}

	return next
}

func Recover(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.Error("request panic",
						slog.Any("panic", recovered),
						slog.String("request_id", requestctx.RequestID(r.Context())),
						slog.String("method", r.Method),
						slog.String("path", r.URL.Path),
					)

					response.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error", nil)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

func RequestContext() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := strings.TrimSpace(r.Header.Get(RequestIDHeader))
			if requestID == "" {
				requestID = newRequestID()
			}

			w.Header().Set(RequestIDHeader, requestID)

			ctx := requestctx.WithRequestID(r.Context(), requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequestLogging(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startedAt := time.Now()
			recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(recorder, r)

			logger.Info("http request",
				slog.String("request_id", requestctx.RequestID(r.Context())),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", recorder.status),
				slog.Duration("duration", time.Since(startedAt)),
			)
		})
	}
}

func RequireAuth(authenticator service.Authenticator) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenant, err := authenticator.Authenticate(
				r.Context(),
				r.Header.Get("X-App-ID"),
				r.Header.Get("Authorization"),
			)
			if err != nil {
				var authErr *service.AuthError
				if errors.As(err, &authErr) {
					response.WriteError(w, http.StatusUnauthorized, authErr.Code, authErr.Message, nil)
					return
				}

				response.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error", nil)
				return
			}

			ctx := requestctx.WithTenant(r.Context(), requestctx.Tenant{
				TenantID: tenant.TenantID,
				AppID:    tenant.AppID,
				Subject:  tenant.Subject,
			})

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func newRequestID() string {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return "fallback-request-id"
	}

	return hex.EncodeToString(buf)
}
