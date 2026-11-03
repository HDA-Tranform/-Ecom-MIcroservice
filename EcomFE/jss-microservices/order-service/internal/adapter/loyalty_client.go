package adapter

import (
    "context"
    "time"

    loyaltypb "github.com/linhhuynhcoding/jss-microservices/rpc/gen/loyalty"
    "go.uber.org/zap"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

// LoyaltyClient wraps the Loyalty gRPC client.  It provides a method
// for applying vouchers to compute discounts on an order.
type LoyaltyClient struct {
    client loyaltypb.LoyaltyClient
    conn   *grpc.ClientConn
    logger *zap.Logger
}

// NewLoyaltyClient dials the loyalty service at the given address.  Call
// Close on the returned client to clean up resources.
func NewLoyaltyClient(addr string, logger *zap.Logger) (*LoyaltyClient, error) {
    conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        return nil, err
    }
    c := loyaltypb.NewLoyaltyClient(conn)
    return &LoyaltyClient{client: c, conn: conn, logger: logger}, nil
}

// Close closes the underlying gRPC connection.
func (c *LoyaltyClient) Close() {
    if c.conn != nil {
        _ = c.conn.Close()
    }
}

// UsingVoucher calls the loyalty service to compute the total discount
// amount for the given vouchers and order totals.  It returns the
// response from the remote service or an error.  A short timeout is
// used to prevent long hangs.
func (c *LoyaltyClient) UsingVoucher(ctx context.Context, req *loyaltypb.UsingVoucherRequest) (*loyaltypb.UsingVoucherResponse, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    return c.client.UsingVoucher(ctx, req)
}