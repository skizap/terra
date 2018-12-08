package client

import (
	"context"
	"time"

	api "github.com/stellarproject/nebula/terra/v1"
)

func (c *Client) Update(manifests []*api.Manifest, force bool) error {
	if _, err := c.client.Update(context.Background(), &api.UpdateRequest{
		ManifestList: &api.ManifestList{
			Manifests: manifests,
			Updated:   time.Now(),
		},
		Force: force,
	}); err != nil {
		return err
	}
	return nil
}
