package discovery

import (
	"context"

	"github.com/google/uuid"
	metadatapb "github.com/neelalala/go-storage/pkg/proto/metadata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	client metadatapb.NodeDiscoveryClient
	conn   *grpc.ClientConn
}

func New(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Client{
		client: metadatapb.NewNodeDiscoveryClient(conn),
		conn:   conn,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Heartbeat(ctx context.Context, id uuid.UUID, addr string) error {
	req := &metadatapb.HeartbeatRequest{
		NodeId:      id.String(),
		NodeAddress: addr,
	}

	_, err := c.client.Heartbeat(ctx, req)
	if err != nil {
		return err
	}

	return nil
}
