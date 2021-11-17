package client

import (
	"context"
	"fmt"
	"net"

	"github.com/chitoku-k/slack-to-ssh/infrastructure/config"
	"github.com/chitoku-k/slack-to-ssh/service"
	"golang.org/x/crypto/ssh"
)

type shellActionExecutor struct {
	Environment  config.Environment
	ClientConfig ssh.ClientConfig
}

func NewShellActionExecutor(environment config.Environment) (service.ActionExecutor, error) {
	signer, err := ssh.ParsePrivateKey(environment.SSH.PrivateKey)
	if err != nil {
		return nil, err
	}

	return &shellActionExecutor{
		Environment: environment,
		ClientConfig: ssh.ClientConfig{
			User:            environment.SSH.Username,
			Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		},
	}, nil
}

func (sae *shellActionExecutor) Do(ctx context.Context, name string) ([]byte, error) {
	var action *service.SlackAction
	for _, v := range sae.Environment.SlackActions {
		if v.Name == name {
			action = &v
			break
		}
	}

	if action == nil {
		return nil, fmt.Errorf("cannot find requested action: %s", name)
	}

	var d net.Dialer
	addr := net.JoinHostPort(sae.Environment.SSH.HostName, sae.Environment.SSH.Port)
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
