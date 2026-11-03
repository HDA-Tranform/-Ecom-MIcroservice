package config

import (
	"fmt"

	"github.com/linhhuynhcoding/jss-microservices/product/consts"
	"github.com/spf13/viper"
)

type Config struct {
	DBSource string `mapstructure:"DB_SOURCE"`

	HttpPort int `mapstructure:"HTTP_PORT"`
	GrpcPort int `mapstructure:"GRPC_PORT"`

	RedisAddress string `mapstructure:"REDIS_ADDRESS"`

	CloudinaryConfig struct {
		ConnectString string `mapstructure:"CLOUDINARY_URL"`
		CloudName     string `mapstructure:"CLOUDINARY_NAME"`
		APIKey        string `mapstructure:"CLOUDINARY_API_KEY"`
		APISecret     string `mapstructure:"CLOUDINARY_API_SECRET"`
		UploadFolder  string `mapstructure:"CLOUDINARY_UPLOAD_FOLDER"`
	} `mapstructure:"CLOUDINARY"`

	UploadFolder string `mapstructure:"UPLOAD_FOLDER"`

	MarketServiceUrl string `mapstructure:"MARKET_SERVICE_URL"`
}

func NewConfig() Config {
	config, err := LoadConfig("./")
	if err != nil {
		panic(err)
	}
	return config
}

func LoadDefaultConfig(cfg *Config) {
	cfg.UploadFolder = consts.DEFAULT_UPLOAD_FOLDER
	cfg.HttpPort = consts.HTTP_PORT
	cfg.GrpcPort = consts.GRPC_PORT
}

func LoadConfig(path string) (config Config, err error) {
	LoadDefaultConfig(&config)

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
