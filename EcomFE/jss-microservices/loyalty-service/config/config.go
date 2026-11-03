package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	DBSource  string `mapstructure:"db_source"`
	MqConnStr string `mapstructure:"mq_conn_str"` // e.g. "amqp://admin:admin@localhost:5672/"
	GrpcPort  string `mapstructure:"grpc_port"`   // default: 50051
	HttpPort  string `mapstructure:"http_port"`   // default: 8080
}

// NewConfig loads configuration from config.yaml and environment variables.
// Environment variables take precedence over file values.
func NewConfig() Config {
	config, err := LoadConfig("./")
	if err != nil {
		// Fall back to environment-only config rather than panicking so the
		// service can still start inside Docker when no config file is present.
		fmt.Printf("config file not found, using environment variables: %v\n", err)
	}
	// Override with environment variables when explicitly set
	applyEnvOverrides(&config)
	return config
}

// applyEnvOverrides reads well-known environment variables and overrides the
// corresponding config fields when the variables are non-empty.
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("DB_SOURCE"); v != "" {
		cfg.DBSource = v
	}
	if v := os.Getenv("MQ_CONN_STR"); v != "" {
		cfg.MqConnStr = v
	}
	if v := os.Getenv("GRPC_PORT"); v != "" {
		cfg.GrpcPort = v
	}
	if v := os.Getenv("HTTP_PORT"); v != "" {
		cfg.HttpPort = v
	}
}

// LoadConfig reads configuration from a YAML file at the given path.
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
