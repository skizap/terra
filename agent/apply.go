package agent

import (
	"context"
	"time"

	ptypes "github.com/gogo/protobuf/types"
	api "github.com/stellarproject/nebula/terra/v1"
)

func (a *Agent) Apply(ctx context.Context, req *api.ApplyRequest) (*ptypes.Empty, error) {
	req.ManifestList.Updated = time.Now()

	if err := a.updateManifestList(req.ManifestList, req.Force); err != nil {
		return empty, err
	}

	return empty, nil
}
