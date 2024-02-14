package handler

import (
	"log"
	"net/http"
)

const (
	somethingWentWrongMsg = "Something went wrong"
	contentTypeJSON       = "application/json;charset=UTF-8"
	contentTypeHeader     = "Content-Type"
)

func (h *Handler) validateTokenMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := h.srv.ValidationBearerToken(r)
		if err != nil {
			handleError(w, http.StatusUnauthorized, err.Error())

			return
		}

		next.ServeHTTP(w, r)
	}
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				log.Printf("Recovery error: %v\n", err)

				handleError(w, http.StatusInternalServerError, somethingWentWrongMsg)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
