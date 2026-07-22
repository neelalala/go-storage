package storage

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/neelalala/go-storage/internal/gateway/domain"
	storagepb "github.com/neelalala/go-storage/pkg/proto/storage"
)

var _ domain.Storage = (*Client)(nil)

type Client struct {
	client storagepb.StorageClient
	conn   *grpc.ClientConn
}

func NewClient(addr string) (*Client, error) {
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

func (c *Client) SaveObject(ctx context.Context, obj domain.Object) (string, error) {
	req := &storagepb.SaveRequest{
		Object: &storagepb.Object{
			Name: obj.Name,
			Data: obj.Data,
		},
	}

	resp, err := c.client.SaveObject(ctx, req)
	if err != nil {
		return "", fmt.Errorf("error saving object: %w", err)
	}

	return resp.GetEtag(), nil
}

func (c *Client) GetObject(ctx context.Context, name string) (domain.Object, error) {
	req := &storagepb.GetRequest{
		Name: name,
	}

	resp, err := c.client.GetObject(ctx, req)
	if err != nil {
		return domain.Object{}, fmt.Errorf("error getting object: %w", err)
	}

	obj := domain.Object{
		Name: name,
		Data: resp.GetData(),
	}

	return obj, nil
}

func (c *Client) DeleteObject(ctx context.Context, name string) error {
	req := &storagepb.DeleteRequest{
		Name: name,
	}

	_, err := c.client.DeleteObject(ctx, req)
	if err != nil {
		return fmt.Errorf("error deleting object: %w", err)
	}

	return nil
}
