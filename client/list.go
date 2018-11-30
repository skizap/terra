package client

import (
	"context"

	api "github.com/stellarproject/nebula/terra/v1"
)

func (c *Client) List() (*api.ManifestList, error) {
	resp, err := c.client.List(context.Background(), &api.ListRequest{})
	if err != nil {
		return nil, err
	}

	return resp.ManifestList, nil
}
