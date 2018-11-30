package agent

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/containerd/typeurl"
	"github.com/gogo/protobuf/types"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
	"github.com/stellarproject/element"
	api "github.com/stellarproject/nebula/terra/v1"
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
}

type AgentConfig struct {
	NodeID                string
	GRPCAddress           string
	ClusterAddress        string
	AdvertiseAddress      string
	Peers                 []string
	Labels                map[string]string
	ConnectionType        string
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
	}
	api.RegisterTerraServer(grpcServer, agent)

	return agent, nil
}

func (a *Agent) Start() error {
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
	t := time.NewTicker(a.clusterAgent.SyncInterval())
	for range t.C {
		peers, err := a.clusterAgent.Peers()
		if err != nil {
			logrus.Error(err)
			continue
		}

		self := a.clusterAgent.Self()
		var updated time.Time
		if self.Payload != nil {
			manifestList, err := parseManifestList(self.Payload)
			if err != nil {
				logrus.Error(err)
				continue
			}

			updated = manifestList.Updated
		}

		for _, peer := range peers {
			if peer.Payload == nil {
				continue
			}
			ml, err := parseManifestList(peer.Payload)
			if err != nil {
				logrus.Errorf("error parsing peer payload: %s", err)
				continue
			}

			if ml.Updated.After(updated) {
				// sync local payload
				a.clusterAgent.Update(peer.Payload)
				logrus.WithFields(logrus.Fields{
					"peer":    peer.ID,
					"updated": ml.Updated,
				}).Info("synchronized payload with peer")
				// TODO: apply assemblies in manifest
			}
		}
	}
}

func parseManifestList(any *types.Any) (*api.ManifestList, error) {
	ml, err := typeurl.UnmarshalAny(any)
	if err != nil {
		return nil, err
	}

	manifestList, ok := ml.(*api.ManifestList)
	if !ok {
		return nil, fmt.Errorf("unexpected local payload %q; expected ManifestList", any.TypeUrl)
	}
	return manifestList, nil
}
