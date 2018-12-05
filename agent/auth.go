package agent

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
)

// RegistryAuth is the base64 encoded credentials for the registry credentials
type RegistryAuth struct {
	Auth string `json:"auth,omitempty"`
}

// DockerConfig is the docker config struct
type DockerConfig struct {
	Auths map[string]RegistryAuth `json:"auths"`
}

func getDockerCredentials(host string) (string, string, error) {
	logrus.WithField("host", host).Debug("checking for registry auth config")
	home, err := homedir.Dir()
	credentialConfig := filepath.Join(home, ".docker", "config.json")
	f, err := os.Open(credentialConfig)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", nil
		}
		return "", "", err
	}
	defer f.Close()

	var cfg DockerConfig
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return "", "", err
	}

	for h, r := range cfg.Auths {
		if h == host {
			creds, err := base64.StdEncoding.DecodeString(r.Auth)
			if err != nil {
				return "", "", err
			}
			parts := strings.SplitN(string(creds), ":", 2)
			logrus.Debugf("using auth for registry %s: user=%s", host, parts[0])
			return parts[0], parts[1], nil
		}
	}

	return "", "", nil
}
