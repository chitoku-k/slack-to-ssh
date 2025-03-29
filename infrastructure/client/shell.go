package client

import (
	"bytes"
	"context"
	"fmt"
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

func NewShellActionExecutor(actions []service.SlackAction, hostname, port, username string, knownHosts []ssh.PublicKey, privateKey ssh.Signer) (service.ActionExecutor, error) {
	var hostKeyCallback ssh.HostKeyCallback
	if knownHosts == nil {
		hostKeyCallback = ssh.InsecureIgnoreHostKey()
	} else {
		hostKeyCallback = verifyHostKey(knownHosts)
	}

	return &shellActionExecutor{
		Actions:  actions,
		HostName: hostname,
		Port:     port,
		ClientConfig: ssh.ClientConfig{
			User:            username,
			Auth:            []ssh.AuthMethod{ssh.PublicKeys(privateKey)},
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
	defer func() {
		_ = client.Close()
	}()

	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to establish a remote session: %w", err)
	}
	defer func() {
		_ = session.Close()
	}()

	output, err := session.CombinedOutput(action.Command)
	if err != nil {
		return nil, fmt.Errorf("failed to get output of the command: %w", err)
	}
	return output, nil
}
