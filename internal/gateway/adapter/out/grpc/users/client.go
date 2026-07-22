package users

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/neelalala/go-storage/internal/gateway/domain"
	userspb "github.com/neelalala/go-storage/pkg/proto/users"
)

type Client struct {
	client userspb.UsersClient
	conn   *grpc.ClientConn
}

func NewClient(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Client{
		client: userspb.NewUsersClient(conn),
		conn:   conn,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) CreateUser(ctx context.Context, name string) (domain.User, error) {
	req := &userspb.CreateUserRequest{
		DisplayName: name,
	}

	resp, err := c.client.CreateUser(ctx, req)
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			return domain.User{}, fmt.Errorf("%w: %s", domain.ErrUserAlreadyExists, name)
		}
		return domain.User{}, err
	}

	userID, err := uuid.Parse(resp.GetId())
	if err != nil {
		return domain.User{}, err
	}

	user := domain.User{
		ID:          userID,
		DisplayName: resp.GetDisplayName(),
	}

	return user, nil
}

func (c *Client) GetUserByName(ctx context.Context, name string) (domain.User, error) {
	req := &userspb.GetUserByNameRequest{
		DisplayName: name,
	}

	resp, err := c.client.GetUserByName(ctx, req)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return domain.User{}, fmt.Errorf("%w: %s", domain.ErrUserNotFound, name)
		}
		return domain.User{}, err
	}

	userID, err := uuid.Parse(resp.GetId())
	if err != nil {
		return domain.User{}, err
	}

	user := domain.User{
		ID:          userID,
		DisplayName: resp.GetDisplayName(),
	}

	return user, nil
}
