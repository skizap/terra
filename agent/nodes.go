package agent

import (
	"context"

	api "github.com/stellarproject/nebula/terra/v1"
)

func (a *Agent) Nodes(ctx context.Context, req *api.NodesRequest) (*api.NodesResponse, error) {
	self := a.clusterAgent.Self()
	peers, err := a.clusterAgent.Peers()
	if err != nil {
		return nil, err
	}

	nodes := []*api.Node{
		{
			ID:      self.ID,
			Address: self.Address,
			Labels:  self.Labels,
		},
	}

	for _, peer := range peers {
		nodes = append(nodes, &api.Node{
			ID:      peer.ID,
			Address: peer.Address,
			Labels:  peer.Labels,
		})
	}

	return &api.NodesResponse{
		Nodes: nodes,
	}, nil
}
