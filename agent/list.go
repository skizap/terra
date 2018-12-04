package agent

import (
	"context"

	api "github.com/stellarproject/nebula/terra/v1"
)

func (a *Agent) List(ctx context.Context, req *api.ListRequest) (*api.ListResponse, error) {
	resp := &api.ListResponse{
		ManifestList: a.manifestList,
	}
	return resp, nil
}
