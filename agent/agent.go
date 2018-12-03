package agent

import (
	"crypto/tls"
	"encoding/json"
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
	bolt "go.etcd.io/bbolt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	bucketState      = "io.stellarproject.terra.v1.state"
	bucketAssemblies = "io.stellarproject.terra.v1.assemblies"
	keyManifestList  = "manifest-list"
)

var (
	empty = &ptypes.Empty{}
)

type status struct {
	mu          *sync.Mutex
	state       api.NodeStatus_Status
	description string
}

func (s *status) State() api.NodeStatus_Status {
	return s.state
}

func (s *status) Description() string {
	return s.description
}

func (s *status) Set(state api.NodeStatus_Status, desc string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = state
	s.description = desc
}

func (s *status) IsUpdating() bool {
	return s.state == api.NodeStatus_UPDATING
}

type Agent struct {
	grpcServer   *grpc.Server
	config       *AgentConfig
	clusterAgent *element.Agent
	mu           *sync.Mutex
	manifestList *api.ManifestList
	db           *bolt.DB
	status       *status
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
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		return nil, err
	}

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

	// database setup
	dbPath := filepath.Join(cfg.DataDir, "terra.db")
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}
	if err := db.Update(func(tx *bolt.Tx) error {
		for _, b := range []string{bucketState, bucketAssemblies} {
			if _, err := tx.CreateBucketIfNotExists([]byte(b)); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer(grpcOpts...)
	agent := &Agent{
		grpcServer:   grpcServer,
		config:       cfg,
		clusterAgent: agt,
		mu:           &sync.Mutex{},
		db:           db,
		status: &status{
			mu:    &sync.Mutex{},
			state: api.NodeStatus_OK,
		},
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

func (a *Agent) Stop() error {
	a.clusterAgent.Shutdown()
	return nil
}

func (a *Agent) sync() {
	t := time.NewTicker(time.Second * 10)
	for range t.C {
		// skip if currently updating
		if a.status.IsUpdating() {
			continue
		}
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
			}).Info("synchronized with peer")
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
	// apply assemblies in manifest
	go func() {
		if err := a.applyManifestList(ml); err != nil {
			logrus.WithError(err).Error("error applying manifest list")
			return
		}
	}()

	return nil
}

func (a *Agent) restoreState() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	var ml *api.ManifestList
	if err := a.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketState))
		v := b.Get([]byte(keyManifestList))

		if v == nil {
			return nil
		}

		if err := json.Unmarshal(v, &ml); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	if ml != nil {
		a.manifestList = ml
		logrus.WithField("updated", a.manifestList.Updated).Debug("restored state")
		if err := a.applyManifestList(ml); err != nil {
			return err
		}
	}

	return nil
}
