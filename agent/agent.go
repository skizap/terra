package agent

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	ptypes "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
	"github.com/stellarproject/element"
	api "github.com/stellarproject/nebula/terra/v1"
	"github.com/stellarproject/terra/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	empty = &ptypes.Empty{}
)

type Agent struct {
	grpcServer   *grpc.Server
	config       *AgentConfig
	clusterAgent *element.Agent
	mu           *sync.Mutex
	manifestList *api.ManifestList
	statePath    string
}

type AgentConfig struct {
	NodeID                string
	GRPCAddress           string
	ClusterAddress        string
	AdvertiseAddress      string
	Peers                 []string
	Labels                map[string]string
	ConnectionType        string
	DataDir               string
	TLSServerCertificate  string
	TLSServerKey          string
	TLSInsecureSkipVerify bool
}

func NewAgent(cfg *AgentConfig) (*Agent, error) {
	grpcOpts := []grpc.ServerOption{}
	if cfg.TLSServerCertificate != "" && cfg.TLSServerKey != "" {
		logrus.WithFields(logrus.Fields{
			"cert": cfg.TLSServerCertificate,
			"key":  cfg.TLSServerKey,
		}).Debug("configuring TLS for GRPC")
		cert, err := tls.LoadX509KeyPair(cfg.TLSServerCertificate, cfg.TLSServerKey)
		if err != nil {
			return nil, err

		}
		creds := credentials.NewTLS(&tls.Config{
			Certificates:       []tls.Certificate{cert},
			ClientAuth:         tls.RequestClientCert,
			InsecureSkipVerify: cfg.TLSInsecureSkipVerify,
		})
		grpcOpts = append(grpcOpts, grpc.Creds(creds))

	}
	agt, err := element.NewAgent(&element.Peer{
		ID:      cfg.NodeID,
		Address: cfg.GRPCAddress,
		Labels:  cfg.Labels,
	}, &element.Config{
		ConnectionType:   cfg.ConnectionType,
		ClusterAddress:   cfg.ClusterAddress,
		AdvertiseAddress: cfg.AdvertiseAddress,
		Peers:            cfg.Peers,
		Debug:            false,
	})
	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer(grpcOpts...)
	agent := &Agent{
		grpcServer:   grpcServer,
		config:       cfg,
		clusterAgent: agt,
		mu:           &sync.Mutex{},
		statePath:    filepath.Join(cfg.DataDir, "state.json"),
	}
	api.RegisterTerraServer(grpcServer, agent)

	return agent, nil
}

func (a *Agent) Start() error {
	if err := a.restoreState(); err != nil {
		return err
	}

	l, err := net.Listen("tcp", a.config.GRPCAddress)
	if err != nil {
		return err
	}

	if err := a.clusterAgent.Start(); err != nil {
		return err
	}

	go a.sync()

	return a.grpcServer.Serve(l)
}

func (a *Agent) sync() {
	t := time.NewTicker(time.Second * 10)
	for range t.C {
		if err := a.syncWithPeers(); err != nil {
			logrus.Error(err)
		}
	}
}

func (a *Agent) syncWithPeers() error {
	peers, err := a.clusterAgent.Peers()
	if err != nil {
		return err
	}

	var updated time.Time
	if a.manifestList != nil {
		updated = a.manifestList.Updated
	}

	logrus.Debugf("current manifest list updated: %s", updated)

	for _, peer := range peers {
		logrus.Debugf("checking peer manifest list: %s", peer.ID)
		// TODO: tls
		c, err := client.NewClient(peer.Address)
		if err != nil {
			logrus.Errorf("error connecting to peer %s: %s", peer.ID, err)
			continue
		}

		ml, err := c.List()
		if err != nil {
			logrus.Errorf("error getting peer manifest list: %s", err)
			c.Close()
			continue
		}

		if ml == nil {
			continue
		}

		logrus.Debugf("peer %s manifest list updated %s", peer.ID, ml.Updated)
		if ml.Updated.After(updated) {
			logrus.Debugf("updating local manifest from peer %s", peer.ID)
			// sync local payload
			if err := a.updateManifestList(ml); err != nil {
				logrus.Errorf("error syncing manifest list with peer: %s", err)
				c.Close()
				continue
			}
			logrus.WithFields(logrus.Fields{
				"peer":    peer.ID,
				"updated": ml.Updated,
			}).Info("synchronized payload with peer")
			// TODO: apply assemblies in manifest
		}
		c.Close()
	}

	return nil
}

func (a *Agent) updateManifestList(ml *api.ManifestList) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// update in memory copy
	a.manifestList = ml

	if err := os.MkdirAll(a.config.DataDir, 0755); err != nil {
		return err
	}

	// persist to disk
	data, err := json.Marshal(ml)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(a.statePath, data, 0644); err != nil {
		return err
	}

	return nil
}

func (a *Agent) restoreState() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, err := os.Stat(a.statePath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	var ml *api.ManifestList
	f, err := os.Open(a.statePath)
	if err != nil {
		return err
	}

	if err := json.NewDecoder(f).Decode(&ml); err != nil {
		return err
	}

	a.manifestList = ml
	logrus.WithField("statePath", a.statePath).Debug("restored state")
	return nil
}
