package config

import (
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	HTTP HTTP `mapstructure:"http"`
	JWT  JWT  `mapstructure:"jwt"`
}

type HTTP struct {
	Port string `mapstructure:"port"`
	// Timeout in seconds
	Timeout int `mapstructure:"tiomeout"`
}

type JWT struct {
	Secret string `mapstructure:"secret"`
	// AccessTokenExpiresIn in hours
	AccessTokenExpiresIn int64 `mapstructure:"access_token_expires_in"`
}

func LoadConfig() (Config, error) {
	_, path, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(path), "../..")

	viper.AddConfigPath(root)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return Config{}, err
	}

	var config Config
	err := viper.Unmarshal(&config)

	return config, err
}
