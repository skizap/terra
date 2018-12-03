package agent

import (
	"context"

	api "github.com/stellarproject/nebula/terra/v1"
	"github.com/stellarproject/terra/client"
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
			Status: &api.NodeStatus{
				Status:      a.status.State(),
				Description: a.status.Description(),
			},
		},
	}

	for _, peer := range peers {
		c, err := client.NewClient(peer.Address)
		if err != nil {
			return nil, err
		}
		nodeStatus, err := c.Status()
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, &api.Node{
			ID:      peer.ID,
			Address: peer.Address,
			Labels:  peer.Labels,
			Status:  nodeStatus,
		})
		c.Close()
	}

	return &api.NodesResponse{
		Nodes: nodes,
	}, nil
}
