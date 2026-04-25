package client

import (
	"context"
	"fmt"

	order "github.com/uncle3dev/velotrax-gateway-go/internal/gen/order"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// OrderClient wraps order service stub.
type OrderClient struct {
	client order.OrderServiceClient
	conn   *grpc.ClientConn
}

// NewOrderClient creates a new OrderClient.
func NewOrderClient(addr string) (*OrderClient, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to order service: %w", err)
	}

	return &OrderClient{
		client: order.NewOrderServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *OrderClient) ListOrders(ctx context.Context, req *order.ListOrdersRequest, authorization string) (*order.ListOrdersResponse, error) {
	return c.client.ListOrders(withAuthorization(ctx, authorization), req)
}

func (c *OrderClient) GetOrder(ctx context.Context, req *order.GetOrderRequest, authorization string) (*order.GetOrderResponse, error) {
	return c.client.GetOrder(withAuthorization(ctx, authorization), req)
}

func (c *OrderClient) GetOrderTracking(ctx context.Context, req *order.GetOrderTrackingRequest, authorization string) (*order.GetOrderTrackingResponse, error) {
	return c.client.GetOrderTracking(withAuthorization(ctx, authorization), req)
}

// Close closes the gRPC connection.
func (c *OrderClient) Close() error {
	return c.conn.Close()
}
