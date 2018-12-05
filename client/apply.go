package client

import (
	"context"
	"time"

	api "github.com/stellarproject/nebula/terra/v1"
)

func (c *Client) Apply(manifests []*api.Manifest, force bool) error {
	if _, err := c.client.Apply(context.Background(), &api.ApplyRequest{
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
