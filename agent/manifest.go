package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/containerd/containerd/archive"
	"github.com/containerd/containerd/archive/compression"
	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/content/local"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	api "github.com/stellarproject/nebula/terra/v1"
	"github.com/stellarproject/terra/client"
	bolt "go.etcd.io/bbolt"
)

func (a *Agent) applyManifestList(ml *api.ManifestList, force bool) error {
	logrus.Debug("applying manifest list")
	a.status.Set(api.NodeStatus_UPDATING, "")
	// check assemblies and install if needed
	for _, manifest := range ml.Manifests {
		if err := a.applyManifest(manifest, force); err != nil {
			a.status.Set(api.NodeStatus_FAILURE, err.Error())
			return err
		}
	}
	a.status.Set(api.NodeStatus_OK, "")

	return nil
}

func (a *Agent) applyManifest(m *api.Manifest, force bool) error {
	matches := false
	// check if node id matches
	if m.NodeID == "" && len(m.Labels) == 0 || a.config.NodeID == m.NodeID {
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

	var errs []string
	for _, assembly := range m.Assemblies {
		logrus.WithField("image", assembly.Image).Info("applying assembly")
		a.status.Set(api.NodeStatus_UPDATING, fmt.Sprintf("applying assembly %s", assembly.Image))
		// apply requires
		for _, req := range assembly.Requires {
			logrus.WithFields(logrus.Fields{
				"image":    assembly.Image,
				"required": req,
			}).Info("applying required assembly")
			output, err := a.applyAssembly(&api.Assembly{Image: req}, force)
			if err != nil {
				logrus.WithError(err).Errorf("error applying required assembly %s: %s", req, string(output))
				errs = append(errs, err.Error())
				continue
			}
		}
		// apply assembly
		output, err := a.applyAssembly(assembly, force)
		if err != nil {
			logrus.WithError(err).Errorf("error applying assembly %s: %s", assembly.Image, string(output))
			errs = append(errs, err.Error())
			continue
		}

		logrus.WithField("assembly", assembly.Image).Info("assembly applied successfully")
	}

	if len(errs) > 0 {
		return fmt.Errorf(strings.Join(errs, ", "))
	}
	return nil
}

func (a *Agent) applyAssembly(assembly *api.Assembly, force bool) ([]byte, error) {
	applied, err := a.assemblyApplied(assembly)
	if err != nil {
		return nil, err
	}
	if applied && !force {
		return nil, err
	}
	tmpdir, err := ioutil.TempDir("", "terra-assembly-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpdir)

	if err := fetchImage(assembly.Image, tmpdir); err != nil {
		return nil, err
	}

	nodePeers := []string{}
	peers, err := a.clusterAgent.Peers()
	if err != nil {
		return nil, err
	}
	for _, peer := range peers {
		nodePeers = append(nodePeers, peer.Address)
	}
	env := os.Environ()
	// add terra env vars
	env = append(env, fmt.Sprintf("TERRA_NODE_ID=%s", a.clusterAgent.Self().ID))
	env = append(env, fmt.Sprintf("TERRA_NODE_ADDR=%s", a.clusterAgent.Self().Address))
	env = append(env, fmt.Sprintf("TERRA_NODE_PEERS=%s", strings.Join(nodePeers, ",")))
	// add parameters
	for k, v := range assembly.Parameters {
		env = append(env, fmt.Sprintf("TERRA_%s=%s", strings.ToUpper(k), v))
	}

	var stdout, stderr bytes.Buffer
	// exec 'install' from package
	cmd := exec.Command("./install")
	cmd.Dir = tmpdir
	cmd.Env = env
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		out := append(stdout.Bytes(), stderr.Bytes()...)
		return nil, errors.Wrap(err, string(out))
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

	authorizer := docker.NewAuthorizer(nil, getDockerCredentials)
	resolver := docker.NewResolver(docker.ResolverOptions{
		Authorizer: authorizer,
	})
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

func (a *Agent) updateManifestList(ml *api.ManifestList, force bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// update in memory copy
	a.manifestList = ml

	// persist to disk
	if err := a.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketState))
		data, err := json.Marshal(ml)
		if err != nil {
			return err
		}
		return b.Put([]byte(keyManifestList), data)
	}); err != nil {
		return err
	}

	logrus.WithField("updated", ml.Updated).Info("updated manifest list")

	//wg := &sync.WaitGroup{}
	// apply assemblies in manifest
	go func() {
		//wg.Add(1)
		//defer wg.Done()
		if err := a.applyManifestList(ml, force); err != nil {
			logrus.WithError(err).Error("error applying manifest list")
			return
		}
	}()
	// if force is specified apply on entire cluster
	if force {
		//wg.Add(1)
		//defer wg.Done()
		peers, err := a.clusterAgent.Peers()
		if err != nil {
			return errors.Wrap(err, "error getting peers to update manifest list")
		}

		for _, peer := range peers {
			go func() {
				//wg.Add(1)
				//defer wg.Done()
				c, err := client.NewClient(peer.Address)
				defer c.Close()
				if err != nil {
					logrus.WithError(err).Errorf("error getting client for peer %s", peer.Address)
					return
				}
				if err := c.Apply(ml.Manifests, true); err != nil {
					logrus.WithError(err).Errorf("error applying manifest list for peer %s", peer.Address)
					return
				}
			}()
		}
	}

	//wg.Wait()

	return nil
}
