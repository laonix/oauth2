package config

import (
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Config struct {
	HTTP HTTP `mapstructure:"http"`
	JWT  JWT  `mapstructure:"jwt"`
	Log  Log  `mapstructure:"log"`
}

type HTTP struct {
	Port    string        `mapstructure:"port"`
	Timeout time.Duration `mapstructure:"timeout"`
}

type JWT struct {
	Secret               string        `mapstructure:"secret"`
	AccessTokenExpiresIn time.Duration `mapstructure:"access_token_expires_in"`
}

type Log struct {
	Level int `mapstructure:"level"`
}

func LoadConfig() (*Config, error) {
	_, path, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(path), "../..")

	viper.AddConfigPath(root)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.Wrap(errors.WithStack(err), "failed to read config")
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, errors.Wrap(errors.WithStack(err), "failed to unmarshal config")
	}

	return &config, nil
}
