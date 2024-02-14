package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/store"

	"oauth/internal/config"
	"oauth/internal/handler"
	"oauth/internal/server"
	"oauth/internal/service/auth"
)

func main() {
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
	clientRepo.Set("client_id", &models.Client{
		ID:     "client_id",
		Secret: "client_secret",
	})

	srv := auth.New(time.Duration(cfg.JWT.AccessTokenExpiresIn), []byte(cfg.JWT.Secret), tokenRepo, clientRepo)

	h := handler.New(srv)
	s := server.New(cfg.HTTP.Port, time.Duration(cfg.HTTP.Timeout), h.Routes())

	go func() {
		if err := s.Run(); err != nil {
			log.Printf("failed to run the http server: %v\n", err.Error())
		}
	}()

	log.Println("server starts")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	if err := s.Shutdown(context.Background()); err != nil {
		log.Printf("failed to shut down the http server: %v\n", err.Error())
	}
}
