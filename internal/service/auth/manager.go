package auth

import (
	"oauth2/internal/config"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/golang-jwt/jwt"
)

func NewManager(cfg *config.Config, tokenRepo oauth2.TokenStore, clientRepo oauth2.ClientStore) oauth2.Manager {
	if cfg == nil {
		return nil
	}

	manager := manage.NewManager()
	managerCfg := &manage.Config{
		AccessTokenExp: cfg.JWT.AccessTokenExpiresIn,
	}

	manager.SetClientTokenCfg(managerCfg)

	manager.MapAccessGenerate(generates.NewJWTAccessGenerate("", []byte(cfg.JWT.Secret), jwt.SigningMethodHS256))

	manager.MapTokenStorage(tokenRepo)
	manager.MapClientStorage(clientRepo)

	return manager
}
