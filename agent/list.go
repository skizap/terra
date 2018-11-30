package agent

import (
	"context"

	api "github.com/stellarproject/nebula/terra/v1"
)

func (a *Agent) List(ctx context.Context, req *api.ListRequest) (*api.ListResponse, error) {
	resp := &api.ListResponse{}
	payload := a.clusterAgent.Self().Payload
	if payload != nil {
		ml, err := parseManifestList(payload)
		if err != nil {
			return nil, err
		}

		resp.ManifestList = ml
	}
	return resp, nil
}
