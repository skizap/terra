package client

import (
	"context"

	api "github.com/stellarproject/nebula/terra/v1"
)

func (c *Client) Status() (*api.NodeStatus, error) {
	resp, err := c.client.Status(context.Background(), &api.StatusRequest{})
	if err != nil {
		return nil, err
	}

	return resp.NodeStatus, nil
}
