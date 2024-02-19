package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"testing"

	"oauth2/internal/config"
	"oauth2/internal/service/auth"

	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/store"
)

const (
	mockClientID     = "client_id"
	mockClientSecret = "client_secret"
)

type GenerateTokenResponse struct {
	Token string `json:"access_token"`
}

func TestGenerateToken(t *testing.T) {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	tokenRepo, err := store.NewMemoryTokenStore()
	if err != nil {
		panic(err)
	}

	clientRepo := store.NewClientStore()

	// mock user
	clientRepo.Set(mockClientID, &models.Client{
		ID:     mockClientID,
		Secret: mockClientSecret,
	})

	srv := auth.NewManager(cfg, tokenRepo, clientRepo)

	httpHandler := New(srv)

	tests := []struct {
		name               string
		w                  *httptest.ResponseRecorder
		prepareRequest     func() *http.Request
		expectedStatusCode int
		accessTokenExists  bool
	}{
		{
			name: "Without client_id and client_secret",
			w:    httptest.NewRecorder(),
			prepareRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "/token?grant_type=client_credentials", nil)
			},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name: "authorization_code grant type",
			w:    httptest.NewRecorder(),
			prepareRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/token?grant_type=authorization_code", nil)
				req.SetBasicAuth(mockClientID, mockClientSecret)

				return req
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "The token should be generated",
			w:    httptest.NewRecorder(),
			prepareRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/token?grant_type=client_credentials", nil)
				req.SetBasicAuth(mockClientID, mockClientSecret)

				return req
			},
			expectedStatusCode: http.StatusOK,
			accessTokenExists:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.prepareRequest()

			httpHandler.generateToken(tt.w, req)

			if tt.w.Code != tt.expectedStatusCode {
				t.Errorf("got status %d but wanted %d\n", tt.w.Code, tt.expectedStatusCode)
			}

			var resp GenerateTokenResponse

			err := json.NewDecoder(tt.w.Body).Decode(&resp)
			if err != nil {
				t.Fatalf("could not decode response: %v\n", err)
			}

			if tt.accessTokenExists && resp.Token == "" {
				t.Errorf("the token must be in the response\n")
			}

			if !tt.accessTokenExists && resp.Token != "" {
				t.Errorf("the token must not be in the response, token: %s\n", resp.Token)
			}
		})
	}
}

func TestValidateTokenMiddleware(t *testing.T) {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	tokenRepo, err := store.NewMemoryTokenStore()
	if err != nil {
		panic(err)
	}

	clientRepo := store.NewClientStore()

	// mock user
	clientRepo.Set(mockClientID, &models.Client{
		ID:     mockClientID,
		Secret: mockClientSecret,
	})

	srv := auth.NewManager(cfg, tokenRepo, clientRepo)

	httpHandler := New(srv)

	req := httptest.NewRequest(http.MethodPost, "/token?grant_type=client_credentials", nil)
	req.SetBasicAuth(mockClientID, mockClientSecret)

	w := httptest.NewRecorder()
	httpHandler.generateToken(w, req)

	var resp GenerateTokenResponse

	err = json.NewDecoder(w.Body).Decode(&resp)
	if err != nil {
		t.Fatalf("could not decode response: %v\n", err)
	}

	validToken := resp.Token

	tests := []struct {
		name               string
		w                  *httptest.ResponseRecorder
		prepareRequest     func() *http.Request
		expectedStatusCode int
	}{
		{
			name: "Without token",
			w:    httptest.NewRecorder(),
			prepareRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "/secure", nil)
			},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name: "Invalid token",
			w:    httptest.NewRecorder(),
			prepareRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/secure", nil)
				req.Header.Add("Authorization", fmt.Sprintln("Bearer mock_token"))

				return req
			},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name: "Valid token",
			w:    httptest.NewRecorder(),
			prepareRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/secure", nil)
				req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))

				return req
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.prepareRequest()

			endpoint := httpHandler.validateTokenMiddleware(http.HandlerFunc(httpHandler.secure))

			endpoint.ServeHTTP(tt.w, req)

			if tt.w.Code != tt.expectedStatusCode {
				t.Errorf("got status %d but wanted %d\n", tt.w.Code, tt.expectedStatusCode)
			}
		})
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	r := mux.NewRouter()
	r.Use(recoveryMiddleware)

	r.HandleFunc("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("panic")
	})

	r.HandleFunc("/recover", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	tests := []struct {
		name               string
		w                  *httptest.ResponseRecorder
		prepareRequest     func() *http.Request
		expectedStatusCode int
	}{
		{
			name: "Panic",
			w:    httptest.NewRecorder(),
			prepareRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/panic", nil)
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name: "Recover",
			w:    httptest.NewRecorder(),
			prepareRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/recover", nil)
			},
			expectedStatusCode: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.prepareRequest()

			r.ServeHTTP(tt.w, req)

			if tt.w.Code != tt.expectedStatusCode {
				t.Errorf("got status %d but wanted %d\n", tt.w.Code, tt.expectedStatusCode)
			}
		})
	}
}
