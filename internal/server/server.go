package server

import (
	"context"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func New(addr string, timeout time.Duration, handler http.Handler) *Server {
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  timeout * time.Second,
		WriteTimeout: timeout * time.Second,
	}

	s := &Server{
		httpServer: httpServer,
	}

	return s
}

// Run starts the HTTP server on the specified port.
func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
