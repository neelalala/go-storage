package storage

import (
	"context"
	"fmt"

	storagepb "github.com/neelalala/go-storage/pkg/proto/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	client storagepb.StorageClient
	conn   *grpc.ClientConn
}

func New(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Client{
		client: storagepb.NewStorageClient(conn),
		conn:   conn,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) DeleteObject(ctx context.Context, path string) error {
	req := &storagepb.DeleteRequest{
		Name: path,
	}

	_, err := c.client.DeleteObject(ctx, req)
	if err != nil {
		return fmt.Errorf("error deleting object: %w", err)
	}

	return nil
}
