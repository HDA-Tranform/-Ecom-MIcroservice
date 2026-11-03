package config

// This package centralises loading configuration for the order service.  It
// reads from the environment using Viper and exposes strongly typed
// configuration values used throughout the service.  If a value is not
// provided via the environment a sensible default is used instead.  See
// .env.example for a list of supported variables.

import (
	"github.com/spf13/viper"
)

// Config holds runtime configuration for the order service.  Ports are
// provided as strings to simplify passing them directly into Listen calls.
// Service addresses should resolve within the Docker network defined in
// docker-compose.
type Config struct {
    HTTPPort           string // Port for the HTTP REST gateway
    GRPCPort           string // Port for the gRPC server
    MongoURI           string // Connection string for MongoDB
    MongoDB            string // Database name within MongoDB
    AuthServiceAddr    string // host:port of the auth service
    ProductServiceAddr string // host:port of the product-customer service
    LoyaltyServiceAddr string // host:port of the loyalty service
    RabbitMQURL        string // Connection string for RabbitMQ
    ExchangeName       string // RabbitMQ exchange for notifications
    PublisherName      string // Name used when publishing messages
}

// Load reads configuration from the environment.  Environment variable
// names are upper‑case and prefixed where appropriate (e.g. ORDER_HTTP_PORT).
// If not set, defaults defined here are used.
func Load() Config {
    // Set default values before reading from the environment.  These
    // correspond to the values used in the repository's docker-compose
    // configuration.
    viper.SetDefault("HTTP_PORT", "8083")
    viper.SetDefault("GRPC_PORT", "50033")
    viper.SetDefault("MONGO_URI", "mongodb://mongo:27017")
    viper.SetDefault("MONGO_DB", "OrderService")
    viper.SetDefault("AUTH_SERVICE_ADDR", "auth-service:50011")
    viper.SetDefault("PRODUCT_SERVICE_ADDR", "product-service:50001")
    viper.SetDefault("LOYALTY_SERVICE_ADDR", "loyalty-service:50051")
    viper.SetDefault("RABBITMQ_URL", "amqp://noti:noti123@rabbitmq:5672/")
    viper.SetDefault("EXCHANGE_NAME", "notifications")
    viper.SetDefault("PUBLISHER_NAME", "order-service")

    viper.AutomaticEnv()

    return Config{
        HTTPPort:           viper.GetString("HTTP_PORT"),
        GRPCPort:           viper.GetString("GRPC_PORT"),
        MongoURI:           viper.GetString("MONGO_URI"),
        MongoDB:            viper.GetString("MONGO_DB"),
        AuthServiceAddr:    viper.GetString("AUTH_SERVICE_ADDR"),
        ProductServiceAddr: viper.GetString("PRODUCT_SERVICE_ADDR"),
        LoyaltyServiceAddr: viper.GetString("LOYALTY_SERVICE_ADDR"),
        RabbitMQURL:        viper.GetString("RABBITMQ_URL"),
        ExchangeName:       viper.GetString("EXCHANGE_NAME"),
        PublisherName:      viper.GetString("PUBLISHER_NAME"),
    }
}