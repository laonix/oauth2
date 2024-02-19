package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"oauth2/internal/handler/response"
	"oauth2/internal/logger"
)

// OAuth2Handler is an interface for handling access token generation and validation.
//
// server.Server implements this interface.
type OAuth2Handler interface {
	HandleTokenRequest(w http.ResponseWriter, r *http.Request) error
	ValidationBearerToken(r *http.Request) (oauth2.TokenInfo, error)
}

// Handler provides routing and requests handling for OAuth2 HTTP server.
type Handler struct {
	srv OAuth2Handler
}

// SecureResponse is a response for secure method.
type SecureResponse struct {
	Message string `json:"message"`
}

// New creates a new instance of Handler.
func New(manager oauth2.Manager) *Handler {
	srvCfg := server.Config{
		TokenType:            "Bearer",
		AllowedResponseTypes: []oauth2.ResponseType{oauth2.Token},
		AllowedGrantTypes:    []oauth2.GrantType{oauth2.ClientCredentials},
	}

	srv := server.NewServer(&srvCfg, manager)

	h := &Handler{
		srv: srv,
	}

	return h
}

// Routes returns the HTTP handler for the OAuth2 server.
//
// It includes the following routes:
//
// - POST /token generates an access token
//
// - POST /secure validates the access token
//
// - GET /health returns the health status of the server
func (h *Handler) Routes() http.Handler {
	r := mux.NewRouter()

	tokenSub := r.PathPrefix("/token").Subrouter()
	tokenSub.Methods(http.MethodPost).HandlerFunc(h.generateToken)
	tokenSub.Use(trackMiddleware, loggingMiddleware, recoveryMiddleware)

	secureSub := r.PathPrefix("/secure").Subrouter()
	secureSub.Methods(http.MethodPost).HandlerFunc(h.secure)
	secureSub.Use(trackMiddleware, loggingMiddleware, recoveryMiddleware, h.validateTokenMiddleware)

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}).Methods(http.MethodGet)

	return r
}

func (h *Handler) generateToken(w http.ResponseWriter, r *http.Request) {
	if err := h.srv.HandleTokenRequest(w, r); err != nil {
		log := logger.WithRequestId(r)
		log.Error().Err(errors.WithStack(err)).Msg("failed to handle token request")

		handleError(w, http.StatusInternalServerError, err.Error())
	}
}

func (h *Handler) secure(w http.ResponseWriter, r *http.Request) {
	resp, err := json.Marshal(&SecureResponse{
		Message: "You have access!",
	})
	if err != nil {
		log := logger.WithRequestId(r)
		log.Error().Err(errors.WithStack(err)).Msg("failed to marshal secure method response")

		handleError(w, http.StatusInternalServerError, err.Error())

		return
	}

	w.Header().Set(contentTypeHeader, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resp)
}

func handleError(w http.ResponseWriter, status int, errMsg string) {
	resp, err := response.NewErrorBody(errMsg)
	if err != nil {
		log := logger.Get()
		log.Error().Err(errors.WithStack(err)).Msg("failed to create error response")

		http.Error(w, errMsg, status)

		return
	}

	w.Header().Set(contentTypeHeader, contentTypeJSON)
	w.WriteHeader(status)
	_, _ = w.Write(resp)
}
