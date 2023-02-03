package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"

	"github.com/chitoku-k/slack-to-ssh/service"
	"golang.org/x/crypto/ssh"
)

type shellActionExecutor struct {
	Actions []service.SlackAction

	HostName     string
	Port         string
	ClientConfig ssh.ClientConfig
}

func parseKnownHosts(knownHosts []byte, hostname string) (publicKey ssh.PublicKey, rest []byte, err error) {
	var marker string
	var hosts []string
	marker, hosts, publicKey, _, rest, err = ssh.ParseKnownHosts(knownHosts)
	if err == io.EOF {
		err = nil
		return
	}
	if err != nil {
		return
	}
	if marker == "revoked" {
		publicKey = nil
		return
	}
	for _, h := range hosts {
		if h == hostname {
			return
		}
	}
	publicKey = nil
	return
}

func verifyHostKey(publicKeys []ssh.PublicKey) ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		for _, knownHostKey := range publicKeys {
			if bytes.Equal(key.Marshal(), knownHostKey.Marshal()) {
				return nil
			}
		}
		return fmt.Errorf("ssh: host key verification failed")
	}
}

func NewShellActionExecutor(actions []service.SlackAction, hostname, port, username string, knownHosts, privateKey []byte) (service.ActionExecutor, error) {
	var hostKeyCallback ssh.HostKeyCallback
	if knownHosts == nil {
		hostKeyCallback = ssh.InsecureIgnoreHostKey()
	} else {
		var publicKeys []ssh.PublicKey
		for knownHosts != nil {
			publicKey, rest, err := parseKnownHosts(knownHosts, hostname)
			if err != nil {
				return nil, fmt.Errorf("failed to parse known hosts: %w", err)
			}
			if publicKey != nil {
				publicKeys = append(publicKeys, publicKey)
			}
			knownHosts = rest
		}
		hostKeyCallback = verifyHostKey(publicKeys)
	}

	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	return &shellActionExecutor{
		Actions:  actions,
		HostName: hostname,
		Port:     port,
		ClientConfig: ssh.ClientConfig{
			User:            username,
			Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
			HostKeyCallback: hostKeyCallback,
		},
	}, nil
}

func (sae *shellActionExecutor) Do(ctx context.Context, name string) ([]byte, error) {
	var action *service.SlackAction
	for _, v := range sae.Actions {
		if v.Name == name {
			action = &v
			break
		}
	}

	if action == nil {
		return nil, fmt.Errorf("cannot find requested action: %s", name)
	}

	var d net.Dialer
	addr := net.JoinHostPort(sae.HostName, sae.Port)
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to remote server: %w", err)
	}

	c, chans, reqs, err := ssh.NewClientConn(conn, addr, &sae.ClientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to establish an SSH connection: %w", err)
	}

	client := ssh.NewClient(c, chans, reqs)
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to establish a remote session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(action.Command)
	if err != nil {
		return nil, fmt.Errorf("failed to get output of the command: %w", err)
	}
	return output, nil
}
