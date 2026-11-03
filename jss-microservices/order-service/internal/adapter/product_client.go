package adapter

import (
    "context"
    "time"

    productpb "github.com/linhhuynhcoding/jss-microservices/rpc/gen/product"
    "go.uber.org/zap"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

// ProductClient wraps the ProductCustomer gRPC client.  It exposes only
// the methods needed by the order service, namely PurchaseProduct.  The
// underlying connection should be closed when no longer in use.
type ProductClient struct {
    client productpb.ProductCustomerClient
    conn   *grpc.ClientConn
    logger *zap.Logger
}

// NewProductClient dials the product service at the given address.  The
// returned client should have Close called when finished.
func NewProductClient(addr string, logger *zap.Logger) (*ProductClient, error) {
    conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        return nil, err
    }
    c := productpb.NewProductCustomerClient(conn)
    return &ProductClient{client: c, conn: conn, logger: logger}, nil
}

// Close closes the underlying gRPC connection.
func (c *ProductClient) Close() {
    if c.conn != nil {
        _ = c.conn.Close()
    }
}

// PurchaseProduct calls the remote PurchaseProduct RPC to update stock
// levels and create order records in the product service.  It returns
// the response or an error if the RPC fails.  A short timeout is used
// to avoid hanging indefinitely.
func (c *ProductClient) PurchaseProduct(ctx context.Context, req *productpb.PurchaseProductRequest) (*productpb.PurchaseProductResponse, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    return c.client.PurchaseProduct(ctx, req)
}