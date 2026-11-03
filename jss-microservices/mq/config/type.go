package config

type RabbitMQConfig struct {
	ConnStr string // amqp://guest:guest@localhost:5672/

	PublisherName  string
	SubscriberName string

	ExchangeType string
	ExchangeName string

	SubscribeKeys []string // just for subscriber. Usage example: ["product.create_product, product.*"]
}
