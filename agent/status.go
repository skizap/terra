package agent

import (
	"context"

	api "github.com/stellarproject/nebula/terra/v1"
)

func (a *Agent) Status(ctx context.Context, req *api.StatusRequest) (*api.StatusResponse, error) {
	return &api.StatusResponse{
		NodeStatus: &api.NodeStatus{
			Status:      a.status.State(),
			Description: a.status.Description(),
		},
	}, nil
}
