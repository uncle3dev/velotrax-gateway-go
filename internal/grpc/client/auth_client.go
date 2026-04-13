package client

import (
	"context"
	"fmt"
	"time"

	auth "github.com/uncle3dev/velotrax-gateway-go/internal/gen/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// AuthClient wraps auth service stub
type AuthClient struct {
	client auth.AuthServiceClient
	conn   *grpc.ClientConn
}

// NewAuthClient creates a new AuthClient
func NewAuthClient(addr string) (*AuthClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service: %w", err)
	}

	client := auth.NewAuthServiceClient(conn)

	return &AuthClient{
		client: client,
		conn:   conn,
	}, nil
}

// Register calls the Register RPC
func (c *AuthClient) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	return c.client.Register(ctx, req)
}

// Login calls the Login RPC
func (c *AuthClient) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	return c.client.Login(ctx, req)
}

// Logout calls the Logout RPC
func (c *AuthClient) Logout(ctx context.Context, req *auth.LogoutRequest) (*auth.LogoutResponse, error) {
	return c.client.Logout(ctx, req)
}

// RefreshToken calls the RefreshToken RPC
func (c *AuthClient) RefreshToken(ctx context.Context, req *auth.RefreshTokenRequest) (*auth.RefreshTokenResponse, error) {
	return c.client.RefreshToken(ctx, req)
}

// Close closes the gRPC connection
func (c *AuthClient) Close() error {
	return c.conn.Close()
}
