package config

import (
	"strings"

	"github.com/spf13/viper"
)

// Config holds all of the configuration values required by the notification
// service.  Values are loaded from a .env file and environment variables.
// See .env.example for an explanation of each field.
type Config struct {
    HTTPPort     string
    GRPCPort     string
    MongoURI     string
    MongoDB      string
    RabbitMQURL  string
    QueueName    string
    ExchangeName string
    LogLevel     string
    JWTSecret    string
    BindingKeys  []string
    SubscriberName  string
}

// Load reads configuration from a .env file in the current working
// directory.  Environment variables with matching names override the
// values found in the file.  If no .env file is present viper will
// silently ignore the missing file and only environment variables are
// used.  It is safe to call Load multiple times – each call will
// return a new Config instance populated with the current environment.
func Load() *Config {
    viper.SetConfigFile(".env.example")
    viper.AutomaticEnv()
    // Ignore file not found errors – in production you may choose
    // to require an explicit config file.
    _ = viper.ReadInConfig()
    keys := strings.Split(viper.GetString("BINDING_KEYS"), ",")
    for i := range keys { keys[i] = strings.TrimSpace(keys[i]) }

    return &Config{
        HTTPPort:     viper.GetString("HTTP_PORT"),
        GRPCPort:     viper.GetString("GRPC_PORT"),
        MongoURI:     viper.GetString("MONGODB_URI"),
        MongoDB:      viper.GetString("MONGO_DB"),
        RabbitMQURL:  viper.GetString("RABBITMQ_URL"),
        QueueName:    viper.GetString("NOTIFICATION_QUEUE_NAME"),
        ExchangeName: viper.GetString("EXCHANGE_NAME"),
        LogLevel:     viper.GetString("LOG_LEVEL"),
        JWTSecret:    viper.GetString("JWT_SECRET"),
        SubscriberName:    viper.GetString("SUBSCRIBER_NAME"),
        BindingKeys:  keys,
        

    }
}