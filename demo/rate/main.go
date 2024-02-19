package main

import (
	"context"
	"fmt"
	"io"
	golog "log"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"time"

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

	log.Info().Any("duration", 10*time.Second).Msg("Start sending requests to oauth2 service...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var reqCount int
rateLoop:
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("timeout")
			break rateLoop
		default:
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost:3000/token?grant_type=client_credentials", nil)
			if err != nil {
				log.Fatal().Err(errors.WithStack(err)).Msg("failed to create request")
				break rateLoop
			}
			req.SetBasicAuth("client_id", "client_secret")

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					log.Info().Msg("timeout")
					break rateLoop
				}

				log.Fatal().Err(errors.WithStack(err)).Msg("failed to send request")
				break rateLoop
			}
			_, _ = io.Copy(io.Discard, res.Body)
			_ = res.Body.Close()

			if res.StatusCode != http.StatusOK {
				log.Warn().Int("status", res.StatusCode).Msg("failed to get token")
			}

			reqCount++
		}
	}

	log.Info().Str("rate", fmt.Sprintf("%d/sec", reqCount/10)).Msg("Finish sending requests to oauth2 service")
}
