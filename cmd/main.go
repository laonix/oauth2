package main

import (
	"context"
	golog "log"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/store"
	"github.com/oklog/run"
	"github.com/pkg/errors"

	"oauth2/internal/config"
	"oauth2/internal/handler"
	"oauth2/internal/logger"
	"oauth2/internal/service/auth"
)

func main() {
	// load config
	cfg, err := config.LoadConfig()
	if err != nil {
		golog.Fatal(errors.Wrap(errors.WithStack(err), "failed to load config"))
	}

	// get logger
	logger.SetLogLevel(cfg.Log.Level)
	log := logger.Get()

	// token store
	tokenRepo, err := store.NewMemoryTokenStore()
	if err != nil {
		log.Fatal().Err(errors.WithStack(err)).Msg("failed to create token store")
	}

	// client store for mock user
	clientRepo := store.NewClientStore()
	if err := clientRepo.Set("client_id", &models.Client{
		ID:     "client_id",
		Secret: "client_secret",
	}); err != nil {
		log.Fatal().Err(errors.WithStack(err)).Msg("failed to set mock client")
	}

	manager := auth.NewManager(cfg, tokenRepo, clientRepo)
	h := handler.New(manager)
	httpServer := &http.Server{
		Addr:         ":" + cfg.HTTP.Port,
		Handler:      h.Routes(),
		ReadTimeout:  cfg.HTTP.Timeout,
		WriteTimeout: cfg.HTTP.Timeout,
	}

	var group run.Group

	// http server
	group.Add(func() error {
		log.Info().Msg("http server listening on port " + cfg.HTTP.Port)

		return httpServer.ListenAndServe()
	}, func(err error) {
		if errors.Is(err, http.ErrServerClosed) {
			log.Info().Msg("http server closed")
		} else {
			log.Error().Err(err).Msg("http server stopped with error")
		}
	})

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// graceful shutdown
	group.Add(func() error {
		<-ctx.Done()

		log.Info().Msg("start graceful shutdown")
		defer log.Info().Msg("graceful shutdown completed")

		return httpServer.Shutdown(context.Background())
	}, func(err error) {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("graceful shutdown interrupted")
		}
	})

	if err := group.Run(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Info().Msg("service stopped")
		} else {
			log.Fatal().Err(err).Msg("service stopped with error")
		}
	}
}
