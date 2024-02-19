package main

import (
	"encoding/json"
	"io"
	golog "log"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"oauth2/internal/logger"
)

func main() {
	_, path, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(path), "../..")

	viper.AddConfigPath(root)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		golog.Fatalf("failed to read config: %+v", errors.WithStack(err))
	}

	log := logger.Get()

	log.Info().Msg("check oauth2 service...")
	for i := 0; i < 5; i++ {
		res, err := http.Get("http://localhost:3000/health")
		if err != nil {
			log.Error().Err(errors.WithStack(err)).Msg("failed to check oauth2 service")
			continue
		}
		_, _ = io.Copy(io.Discard, res.Body)
		_ = res.Body.Close()

		if res.StatusCode == http.StatusOK {
			log.Info().Msg("oauth2 service is running")
			break
		}
	}

	log.Info().Msg("1. Get access token")
	req, err := http.NewRequest("POST", "http://localhost:3000/token?grant_type=client_credentials", nil)
	if err != nil {
		log.Fatal().Err(errors.WithStack(err)).Msg("failed to create request for retrieving access token")
	}
	req.SetBasicAuth("client_id", "client_secret")

	log.Info().Msg("Sending request to retrieve access token")
	tokenRes, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal().Err(errors.WithStack(err)).Msg("failed to retrieve access token")
	}
	defer tokenRes.Body.Close()

	log.Info().Msg("Decoding access token response")
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}
	if err := json.NewDecoder(tokenRes.Body).Decode(&tokenResp); err != nil {
		log.Fatal().Err(errors.WithStack(err)).Msg("failed to decode access token response")
	}

	log.Info().
		Any("access_token", tokenResp.AccessToken).
		Any("expires_in", tokenResp.ExpiresIn).
		Any("token_type", tokenResp.TokenType).
		Msg("access token retrieved")

	log.Info().Msg("2. Validate access token")
	req, err = http.NewRequest("POST", "http://localhost:3000/secure", nil)
	if err != nil {
		log.Fatal().Err(errors.WithStack(err)).Msg("failed to create request for validating access token")
	}
	req.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)

	log.Info().Msg("Sending request to validate access token")
	secureRes, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal().Err(errors.WithStack(err)).Msg("failed to validate access token")
	}
	defer secureRes.Body.Close()
	_, _ = io.Copy(io.Discard, secureRes.Body)

	if secureRes.StatusCode != http.StatusOK {
		log.Warn().Msg("access token is invalid")
	} else {
		log.Info().Msg("access token is valid")
	}
}
