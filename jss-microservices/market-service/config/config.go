package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/viper"
)

type Config struct {
	DBSource string `mapstructure:"DB_SOURCE"`

	HttpPort int `mapstructure:"HTTP_PORT"`
	GrpcPort int `mapstructure:"GRPC_PORT"`

	RedisAddress string `mapstructure:"REDIS_ADDRESS"`
}

func NewConfig() Config {
	config, err := LoadConfig("./")
	if err != nil {
		fmt.Printf("[market-service] warning: config file not found, using env/defaults: %v\n", err)
	}
	applyEnvOverrides(&config)
	return config
}

// applyEnvOverrides reads environment variables and overrides any zero-value
// fields loaded from the config file.
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("DB_SOURCE"); v != "" {
		cfg.DBSource = v
	}
	if v := os.Getenv("GRPC_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.GrpcPort = p
		}
	}
	if v := os.Getenv("HTTP_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.HttpPort = p
		}
	}
	if v := os.Getenv("REDIS_ADDRESS"); v != "" {
		cfg.RedisAddress = v
	}
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		fmt.Printf("looking for config in: %v\n", viper.ConfigFileUsed())
		return
	}

	err = viper.Unmarshal(&config)
	return
}
