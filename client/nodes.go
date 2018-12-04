package client

import (
	"context"

	api "github.com/stellarproject/nebula/terra/v1"
)

func (c *Client) Nodes() ([]*api.Node, error) {
	resp, err := c.client.Nodes(context.Background(), &api.NodesRequest{})
	if err != nil {
		return nil, err
	}

	return resp.Nodes, nil
}
