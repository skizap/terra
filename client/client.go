package client

import (
	"context"
	"time"

	api "github.com/stellarproject/nebula/terra/v1"
	"google.golang.org/grpc"
)

type Client struct {
	conn   *grpc.ClientConn
	client api.TerraClient
}

func NewClient(addr string, opts ...grpc.DialOption) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if len(opts) == 0 {
		opts = []grpc.DialOption{
			grpc.WithInsecure(),
		}
	}

	opts = append(opts, grpc.WithWaitForHandshake())
	c, err := grpc.DialContext(ctx,
		addr,
		opts...,
	)
	if err != nil {
		return nil, err

	}

	return &Client{
		conn:   c,
		client: api.NewTerraClient(c),
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
