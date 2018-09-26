package installer

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/containerd/containerd/archive"
	"github.com/containerd/containerd/archive/compression"
	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/content/local"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// FetchImage retrieves image content from a remote registry
func FetchImage(imageName, dest string) error {
	if _, err := os.Stat(dest); err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		if err := os.MkdirAll(dest, 0755); err != nil {
			return err
		}
	}
	tmpContent, err := ioutil.TempDir("", "terra-content-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpContent)

	cs, err := local.NewStore(tmpContent)
	if err != nil {
		return err
	}
	resolver := docker.NewResolver(docker.ResolverOptions{})
	ctx := context.Background()

	name, desc, err := resolver.Resolve(ctx, imageName)
	if err != nil {
		return err
	}
	fetcher, err := resolver.Fetcher(ctx, name)
	if err != nil {
		return err
	}
	r, err := fetcher.Fetch(ctx, desc)
	if err != nil {
		return err
	}
	defer r.Close()

	childrenHandler := images.ChildrenHandler(cs)
	h := images.Handlers(remotes.FetchHandler(cs, fetcher), childrenHandler)
	if err := images.Dispatch(ctx, h, desc); err != nil {
		return err
	}

	if err := cs.Walk(ctx, func(info content.Info) error {
		desc := ocispec.Descriptor{
			Digest: info.Digest,
		}
		ra, err := cs.ReaderAt(ctx, desc)
		if err != nil {
			return err
		}
		cr := content.NewReader(ra)
		r, err := compression.DecompressStream(cr)
		if err != nil {
			return err
		}
		defer r.Close()
		if r.(compression.DecompressReadCloser).GetCompression() == compression.Uncompressed {
			return nil
		}
		if _, err := archive.Apply(ctx, dest, r); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}
