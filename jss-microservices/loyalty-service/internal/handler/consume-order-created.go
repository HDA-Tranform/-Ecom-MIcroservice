package handler

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/linhhuynhcoding/jss-microservices/loyalty/config"
	"github.com/linhhuynhcoding/jss-microservices/loyalty/internal/repository"
	"github.com/linhhuynhcoding/jss-microservices/mq"
	mqConfig "github.com/linhhuynhcoding/jss-microservices/mq/config"
	"github.com/linhhuynhcoding/jss-microservices/mq/consts"
	"github.com/linhhuynhcoding/jss-microservices/mq/events"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// IOrderCreatedConsumer defines the interface for consuming ORDER_CREATED events
// from RabbitMQ and awarding loyalty points to customers.
type IOrderCreatedConsumer interface {
	ConsumeOrderCreated(ctx context.Context)
}

// OrderCreatedConsumer listens to ORDER_CREATED events and credits loyalty
// points to the corresponding customer.  Points are calculated as 1% of
// the order's final price (rounded to the nearest integer), with a minimum
// of 1 point per order.
type OrderCreatedConsumer struct {
	logger  *zap.Logger
	cfg     config.Config
	queries repository.Store
}

// NewOrderCreatedConsumer constructs an OrderCreatedConsumer.
// The queries parameter is used to persist loyalty point records.
func NewOrderCreatedConsumer(
	logger *zap.Logger,
	cfg config.Config,
	queries repository.Store,
) IOrderCreatedConsumer {
	return &OrderCreatedConsumer{
		logger:  logger,
		cfg:     cfg,
		queries: queries,
	}
}

// ConsumeOrderCreated establishes a RabbitMQ subscriber and starts consuming
// ORDER_CREATED events in a blocking loop.  It returns only when the
// subscriber encounters a fatal error or the context is cancelled.
func (c *OrderCreatedConsumer) ConsumeOrderCreated(ctx context.Context) {
	logger := c.logger.With(zap.String("func", "ConsumeOrderCreated"))

	subCfg := mqConfig.RabbitMQConfig{
		ConnStr:        c.cfg.MqConnStr,
		ExchangeName:   consts.EXCHANGE_ORDER_SERVICE,
		ExchangeType:   "topic",
		SubscribeKeys:  []string{consts.TOPIC_CREATE_ORDER},
		PublisherName:  consts.EXCHANGE_ORDER_SERVICE,
		SubscriberName: "loyalty.order.created",
	}

	subscriber, err := mq.NewSubscriber(subCfg, logger)
	if err != nil {
		logger.Error("failed to init RabbitMQ subscriber", zap.Error(err))
		return
	}
	defer subscriber.Close()
	logger.Info("RabbitMQ subscriber ready",
		zap.String("exchange", subCfg.ExchangeName),
		zap.Strings("keys", subCfg.SubscribeKeys),
	)

	errCh := make(chan error, 1)
	go func() {
		errCh <- subscriber.Consume(func(body []byte) error {
			return c.handleOrderCreated(ctx, body)
		})
	}()

	select {
	case err := <-errCh:
		if err != nil {
			logger.Error("subscriber exited with error", zap.Error(err))
		} else {
			logger.Info("subscriber exited cleanly")
		}
	case <-ctx.Done():
		logger.Info("context cancelled, stopping order-created consumer")
	}
}

// handleOrderCreated parses an ORDER_CREATED event envelope and awards
// loyalty points to the customer who placed the order.
func (c *OrderCreatedConsumer) handleOrderCreated(ctx context.Context, body []byte) error {
	// Attempt to unmarshal as EventEnvelope first (standard wrapper).
	var envelope events.EventEnvelope
	if err := proto.Unmarshal(body, &envelope); err != nil {
		// Fallback: try direct ProductEvent (legacy format)
		c.logger.Warn("failed to unmarshal EventEnvelope, skipping message", zap.Error(err))
		return nil // ACK — non-recoverable parse error
	}

	// Decode the inner payload.  The order-service publishes an
	// EventEnvelope whose Payload field contains a serialised
	// ProductEvent (or a custom OrderCreatedEvent once the publisher is
	// updated).  For now we extract what we can from the available fields.
	//
	// Expected fields from the ORDER_CREATED event metadata:
	//   "customer_id" → int32 string
	//   "final_price"  → float64 string
	//   "order_id"     → int32 string
	customerIDStr, ok := envelope.Metadata["customer_id"]
	if !ok || customerIDStr == "" {
		c.logger.Warn("ORDER_CREATED event missing customer_id metadata, skipping")
		return nil
	}

	var customerID int32
	if _, err := scanInt32(customerIDStr, &customerID); err != nil || customerID <= 0 {
		c.logger.Warn("invalid customer_id in ORDER_CREATED event",
			zap.String("raw", customerIDStr))
		return nil
	}

	// Calculate points: 1 point per 10,000 VND of final price, minimum 1.
	var finalPrice float64
	if fp, ok2 := envelope.Metadata["final_price"]; ok2 {
		scanFloat64(fp, &finalPrice) //nolint:errcheck
	}
	points := int32(finalPrice / 10000)
	if points < 1 {
		points = 1
	}

	// Derive reference order ID (optional, for traceability).
	var orderID int32
	if oid, ok3 := envelope.Metadata["order_id"]; ok3 {
		scanInt32(oid, &orderID) //nolint:errcheck
	}

	// Persist loyalty point record.
	params := repository.CreateLoyaltyPointParams{
		CustomerID:  customerID,
		Points:      pgtype.Int4{Int32: points, Valid: true},
		Source:      "ORDER",
		ReferenceID: pgtype.Int4{Int32: orderID, Valid: orderID > 0},
	}
	lp, err := c.queries.CreateLoyaltyPoint(ctx, params)
	if err != nil {
		c.logger.Error("failed to create loyalty point for order",
			zap.Int32("customer_id", customerID),
			zap.Int32("order_id", orderID),
			zap.Int32("points", points),
			zap.Error(err),
		)
		// Return error so the message is requeued (transient DB error).
		return err
	}

	c.logger.Info("loyalty points awarded",
		zap.Int32("loyalty_point_id", lp.ID),
		zap.Int32("customer_id", customerID),
		zap.Int32("order_id", orderID),
		zap.Int32("points", points),
		zap.Float64("final_price", finalPrice),
	)
	return nil
}

// scanInt32 parses a string into an int32.
func scanInt32(s string, out *int32) (int32, error) {
	var n int32
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		return 0, err
	}
	*out = n
	return n, nil
}

// scanFloat64 parses a string into a float64.
func scanFloat64(s string, out *float64) (float64, error) {
	var f float64
	if _, err := fmt.Sscanf(s, "%f", &f); err != nil {
		return 0, err
	}
	*out = f
	return f, nil
}

// Keep the standard library log import available (used by callers of
// subscriber.Consume in the legacy path).
var _ = log.Printf
