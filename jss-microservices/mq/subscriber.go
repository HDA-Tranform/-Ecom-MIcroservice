package mq

import (
	"context"
	"fmt"
	"time"

	"github.com/linhhuynhcoding/jss-microservices/mq/config"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type Subscriber struct {
	pubConfig config.RabbitMQConfig
	conn      *amqp.Connection
	ch        *amqp.Channel
	q         amqp.Queue

	ctx    context.Context
	cancel context.CancelFunc
	logger *zap.Logger
}

func NewSubscriber(
	cfg config.RabbitMQConfig,
	log *zap.Logger,
) (*Subscriber, error) {
	ctx, cancel := context.WithCancel(context.Background())
	logger := log.With(zap.String("Subscriber", cfg.SubscriberName))

	conn, err := amqp.Dial(cfg.ConnStr)
	if err != nil {
		logger.Error("Failed to connect mq", zap.Error(err))
		defer cancel()
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		logger.Error("Failed to init mq channel", zap.Error(err))
		defer cancel()
		return nil, err
	}

	err = ch.ExchangeDeclare(
		cfg.ExchangeName, // name
		cfg.ExchangeType, // type
		true,             // durable
		false,            // auto-deleted
		false,            // internal
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		logger.Error("Failed to init mq exchange", zap.Error(err))
		defer cancel()
		return nil, err
	}

	q, err := ch.QueueDeclare(
		"",    // name (should be empty to auto-generate)
		true,  // durable
		true,  // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		logger.Error("Failed to init mq queue", zap.Error(err))
		defer cancel()
		return nil, err
	}

	return &Subscriber{
		pubConfig: cfg,
		conn:      conn,
		ch:        ch,
		q:         q,
		ctx:       ctx,
		cancel:    cancel,
		logger:    logger,
	}, nil
}

func (s *Subscriber) Consume(handler func([]byte) error) error {
	var (
		err error
	)
	for _, key := range s.pubConfig.SubscribeKeys {
		err = s.ch.QueueBind(
			s.q.Name,                 // queue name
			key,                      // routing key
			s.pubConfig.ExchangeName, // exchange
			false,
			nil,
		)
		if err != nil {
			s.logger.Error("failed to binding key", zap.Any("key", key), zap.Error(err))
			// ignore err
		}
	}

	// Set QoS to control message prefetch
	err = s.ch.Qos(
		10,    // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := s.ch.Consume(
		s.q.Name, // queue
		"",       // consumer
		false,    // auto-ack
		false,    // exclusive
		false,    // no-local
		false,    // no-wait
		nil,      // args
	)
	if err != nil {
		s.logger.Error("failed to consume", zap.Error(err))
		return err
	}

	s.logger.Info("Started consuming messages", zap.String("queue", s.q.Name))
	for {
		select {
		case <-s.ctx.Done():
			return s.ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				s.logger.Warn("Message channel closed")
			} else {
				go s.handleMessage(msg, handler)
			}
		}

	}
}

func (s *Subscriber) Close() {
	if s.ch != nil {
		err := s.ch.Close()
		if err != nil {
			s.logger.Warn("failed to closing channel")
		}
	}

	if s.conn != nil {
		err := s.conn.Close()
		if err != nil {
			s.logger.Warn("failed to closing connection")
		}
	}
}

func (s *Subscriber) handleMessage(delivery amqp.Delivery, handler func([]byte) error) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("Panic in message handler",
				zap.Any("panic", r),
				zap.String("routingKey", delivery.RoutingKey))

			// Reject message and requeue on panic
			if err := delivery.Nack(false, true); err != nil {
				s.logger.Error("Failed to nack message after panic", zap.Error(err))
			}
		}
	}()

	// Process message with timeout context
	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- handler(delivery.Body)
	}()

	select {
	case <-ctx.Done():
		s.logger.Error("Message handler timeout",
			zap.String("routingKey", delivery.RoutingKey))

		// Reject and requeue on timeout
		if err := delivery.Nack(false, true); err != nil {
			s.logger.Error("Failed to nack message after timeout", zap.Error(err))
		}

	case err := <-done:
		if err != nil {
			s.logger.Error("Message handler failed",
				zap.Error(err),
				zap.String("routingKey", delivery.RoutingKey))

			// Reject and requeue on handler error
			if nackErr := delivery.Nack(false, true); nackErr != nil {
				s.logger.Error("Failed to nack message after handler error", zap.Error(nackErr))
			}
		} else {
			// Acknowledge successful processing
			if ackErr := delivery.Ack(false); ackErr != nil {
				s.logger.Error("Failed to ack message", zap.Error(ackErr))
			}
		}
	}
}
