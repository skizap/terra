package agent

import (
	"context"
	"time"

	ptypes "github.com/gogo/protobuf/types"
	api "github.com/stellarproject/nebula/terra/v1"
)

func (a *Agent) Update(ctx context.Context, req *api.UpdateRequest) (*ptypes.Empty, error) {
	req.ManifestList.Updated = time.Now()

	if err := a.updateManifestList(req.ManifestList, req.Force); err != nil {
		return empty, err
	}

	return empty, nil
}
