package agent

import (
	"context"

	"github.com/containerd/typeurl"
	ptypes "github.com/gogo/protobuf/types"
	api "github.com/stellarproject/nebula/terra/v1"
)

func (a *Agent) Apply(ctx context.Context, req *api.ApplyRequest) (*ptypes.Empty, error) {
	any, err := typeurl.MarshalAny(req.ManifestList)
	if err != nil {
		return empty, err
	}

	a.clusterAgent.Update(any)

	return empty, nil
}
