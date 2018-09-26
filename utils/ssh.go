package utils

import (
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// SSHClient is a client to connect and interact with a remote SSH server
type SSHClient struct {
	client *ssh.Client
}

// NewSSHClient returns a new SSHClient to use with a remote host
func NewSSHClient(user, host string) (*SSHClient, error) {
	authMethods, err := getAuthMethods()
	if err != nil {
		return nil, err
	}
	config := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	c, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to connect to %s", host)
	}

	return &SSHClient{
		client: c,
	}, nil
}

// Client returns the underlying `ssh.Client`
func (c *SSHClient) Client() *ssh.Client {
	return c.client
}

// Exec runs a command on the remote host and returns the output
func (c *SSHClient) Exec(cmd string) ([]byte, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	sout, err := session.StdoutPipe()
	if err != nil {
		return nil, err
	}

	serr, err := session.StderrPipe()
	if err != nil {
		return nil, err
	}
	go func() {
		io.Copy(os.Stdout, sout)
	}()

	go func() {
		io.Copy(os.Stderr, serr)
	}()

	if err := session.Start(cmd); err != nil {
		return nil, err
	}

	if err := session.Wait(); err != nil {
		return nil, err
	}

	return nil, nil
}

// Close closes the underlying connection
func (c *SSHClient) Close() error {
	return c.client.Close()
}

func getAuthMethods() ([]ssh.AuthMethod, error) {
	auth := []ssh.AuthMethod{}
	// support ssh agent
	if socket := os.Getenv("SSH_AUTH_SOCK"); socket != "" {
		conn, err := net.Dial("unix", socket)
		if err != nil {
			return nil, err
		}
		agentClient := agent.NewClient(conn)
		auth = append(auth, ssh.PublicKeysCallback(agentClient.Signers))
	}

	// load user id_rsa
	home, err := homedir.Dir()
	if err != nil {
		return nil, err
	}
	privateKey := filepath.Join(home, ".ssh", "id_rsa")
	if _, err := os.Stat(privateKey); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		return auth, nil
	}

	key, err := ioutil.ReadFile(privateKey)
	if err != nil {
		return nil, err
	}
	if key != nil {
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, err
		}

		auth = append(auth, ssh.PublicKeys(signer))
	}

	return auth, nil
}
