package auth

import (
	"time"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/golang-jwt/jwt"
)

func New(accessTokenExp time.Duration, jwtSecret []byte, tokenRepo oauth2.TokenStore, clientRepo oauth2.ClientStore) oauth2.Manager {
	manager := manage.NewManager()
	cfg := &manage.Config{
		AccessTokenExp: accessTokenExp * time.Hour,
	}

	manager.SetClientTokenCfg(cfg)

	manager.MapAccessGenerate(generates.NewJWTAccessGenerate("", jwtSecret, jwt.SigningMethodPS256))

	manager.MapTokenStorage(tokenRepo)
	manager.MapClientStorage(clientRepo)

	return manager
}
