package agent

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/containerd/containerd/archive"
	"github.com/containerd/containerd/archive/compression"
	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/content/local"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
	api "github.com/stellarproject/nebula/terra/v1"
	bolt "go.etcd.io/bbolt"
)

func (a *Agent) applyManifestList(ml *api.ManifestList) error {
	a.updating = true
	defer func() {
		a.updating = false
	}()
	// check assemblies and install if needed
	for _, manifest := range ml.Manifests {
		if err := a.applyManifest(manifest); err != nil {
			logrus.WithError(err).Error("error applying manifest")
			continue
		}
	}

	return nil
}

func (a *Agent) applyManifest(m *api.Manifest) error {
	matches := false
	// check if node id matches
	if m.NodeID == "" || a.config.NodeID == m.NodeID {
		matches = true
	}
	// check labels
	for k, v := range m.Labels {
		if x, ok := a.config.Labels[k]; ok {
			if x == "" || x == v {
				matches = true
				break
			}
		}
	}

	if !matches {
		return nil
	}

	for _, assembly := range m.Assemblies {
		logrus.WithField("image", assembly.Image).Info("applying assembly")
		if output, err := a.applyAssembly(assembly); err != nil {
			logrus.WithError(err).Errorf("error applying assembly %s: %s", assembly.Image, string(output))
			continue
		}

		logrus.WithField("assembly", assembly.Image).Info("assembly applied successfully")
	}

	return nil
}

func (a *Agent) applyAssembly(assembly *api.Assembly) ([]byte, error) {
	exists, err := a.assemblyApplied(assembly)
	if err != nil {
		return nil, err
	}
	if exists {
		logrus.WithFields(logrus.Fields{
			"assembly": assembly.Image,
		}).Debug("assembly already applied")
		return nil, nil
	}
	tmpdir, err := ioutil.TempDir("", "terra-assembly-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpdir)

	if err := fetchImage(assembly.Image, tmpdir); err != nil {
		return nil, err
	}

	var stdout, stderr bytes.Buffer
	// exec 'install' from package
	cmd := exec.Command("./install")
	cmd.Dir = tmpdir
	cmd.Env = os.Environ()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	output := append(stdout.Bytes(), stderr.Bytes()...)

	// update db
	if err := a.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketAssemblies))
		return b.Put([]byte(assembly.Image), output)
	}); err != nil {
		return output, err
	}

	return output, nil
}

func (a *Agent) assemblyApplied(assembly *api.Assembly) (bool, error) {
	applied := false
	if err := a.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketAssemblies))
		v := b.Get([]byte(assembly.Image))
		if v != nil {
			applied = true
		}
		return nil
	}); err != nil {
		return false, err
	}

	return applied, nil
}

func fetchImage(imageName, dest string) error {
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
