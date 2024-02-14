package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"oauth/internal/handler/response"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/server"
)

type OauthHandler interface {
	HandleTokenRequest(w http.ResponseWriter, r *http.Request) error
	ValidationBearerToken(r *http.Request) (oauth2.TokenInfo, error)
}

type Handler struct {
	srv OauthHandler
}

type SecureResponse struct {
	Message string `json:"message"`
}

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

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/token", h.generateToken)
	mux.HandleFunc("/secure", h.validateTokenMiddleware(h.secure))

	router := recoveryMiddleware(mux)

	return router
}

func (h *Handler) generateToken(w http.ResponseWriter, r *http.Request) {
	err := h.srv.HandleTokenRequest(w, r)
	if err != nil {
		handleError(w, http.StatusInternalServerError, err.Error())
	}
}

func (h *Handler) secure(w http.ResponseWriter, r *http.Request) {
	msg := SecureResponse{
		Message: "You have access!",
	}

	resp, err := json.Marshal(&msg)
	if err != nil {
		log.Printf("secure Marshal error: %v\n", err)

		handleError(w, http.StatusInternalServerError, err.Error())

		return
	}

	w.Header().Set(contentTypeHeader, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func handleError(w http.ResponseWriter, status int, errMsg string) {
	resp, err := response.NewErrorBody(errMsg)
	if err != nil {
		log.Printf("handleError NewErrorBody error: %v\n", err.Error())

		http.Error(w, errMsg, status)

		return
	}

	w.Header().Set(contentTypeHeader, contentTypeJSON)
	w.WriteHeader(status)
	w.Write(resp)
}
