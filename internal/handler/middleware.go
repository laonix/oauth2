package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"oauth2/internal/logger"
)

const (
	somethingWentWrongMsg = "Something went wrong"
	contentTypeJSON       = "application/json;charset=UTF-8"
	contentTypeHeader     = "Content-Type"
)

func (h *Handler) validateTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := h.srv.ValidationBearerToken(r)
		if err != nil {
			handleError(w, http.StatusUnauthorized, err.Error())

			return
		}

		next.ServeHTTP(w, r)
	})
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				log := logger.WithRequestId(r)
				log.Warn().Any("panic", fmt.Sprintf("%+v", err)).Msg("recovered from panic")

				handleError(w, http.StatusInternalServerError, somethingWentWrongMsg)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// trackMiddleware tracks the request by adding a unique request ID to the request context and response headers.
func trackMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestId := uuid.New().String()

		r = r.WithContext(context.WithValue(r.Context(), "X-Request-ID", requestId))
		w.Header().Add("X-Request-ID", requestId)

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs the incoming request.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now().UTC()

		log := logger.WithRequestId(r)

		log.
			Info().
			Str("method", r.Method).
			Str("url", r.URL.RequestURI()).
			Msg("incoming request")

		defer func() {
			log.
				Info().
				Str("method", r.Method).
				Str("url", r.URL.RequestURI()).
				Dur("elapsed_ms", time.Since(start)).
				Msg("request served")
		}()

		next.ServeHTTP(w, r)
	})
}
