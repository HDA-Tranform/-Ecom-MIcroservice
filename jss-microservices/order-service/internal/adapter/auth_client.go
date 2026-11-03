package adapter

// This package provides thin wrappers around gRPC clients for interacting
// with other microservices.  By encapsulating client creation and method
// calls here, the service layer remains clean and easy to test.  Each
// adapter lazily dials the underlying service when constructed and
// exposes a small interface tailored to the order service's needs.

import (
    "context"
    "time"

    authpb "github.com/linhhuynhcoding/jss-microservices/rpc/gen/auth"
    "go.uber.org/zap"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

// AuthClient wraps the gRPC client for the AuthService.  It exposes a
// Validate method which validates an access token and returns the user
// identifier and role.  The underlying gRPC connection is established
// when NewAuthClient is called.
type AuthClient struct {
    client authpb.AuthServiceClient
    conn   *grpc.ClientConn
    logger *zap.Logger
}

// NewAuthClient dials the auth service at the given address.  Call
// Close when finished using the client to close the underlying
// connection.
func NewAuthClient(addr string, logger *zap.Logger) (*AuthClient, error) {
    // gRPC dial with insecure credentials because communication occurs
    // inside the trusted docker network.  In a production setting
    // mTLS would be preferred.
    conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        return nil, err
    }
    client := authpb.NewAuthServiceClient(conn)
    return &AuthClient{client: client, conn: conn, logger: logger}, nil
}

// Close closes the underlying gRPC connection.
func (c *AuthClient) Close() {
    if c.conn != nil {
        _ = c.conn.Close()
    }
}

// Validate verifies the provided JWT access token by calling the
// AuthService.ValidateToken RPC.  It returns whether the token is
// valid, the associated user ID and role.  On error the boolean will
// be false and the user ID and role empty.
func (c *AuthClient) Validate(ctx context.Context, token string) (bool, string, string, error) {
    // Use a short timeout to avoid blocking indefinitely if the auth
    // service is unavailable.
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    resp, err := c.client.ValidateToken(ctx, &authpb.ValidateTokenRequest{AccessToken: token})
    if err != nil {
        return false, "", "", err
    }
    return resp.GetIsValid(), resp.GetUserId(), resp.GetRole(), nil
}