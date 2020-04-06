package client

import (
	"github.com/chitoku-k/slack-to-ssh/infrastructure/config"
	"github.com/chitoku-k/slack-to-ssh/service"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

type shellActionExecutor struct {
	Environment  config.Environment
	ClientConfig ssh.ClientConfig
}

func NewShellActionExecutor(environment config.Environment) (service.ActionExecutor, error) {
	signer, err := ssh.ParsePrivateKey([]byte(environment.SSH.PrivateKey))
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

func (sae *shellActionExecutor) Do(name string) ([]byte, error) {
	var action *service.SlackAction
	for _, v := range sae.Environment.SlackActions {
		if v.Name == name {
			action = &v
			break
		}
	}

	if action == nil {
		return nil, errors.New("cannot find requested action: " + name)
	}

	client, err := ssh.Dial(
		"tcp",
		sae.Environment.SSH.HostName+":"+sae.Environment.SSH.Port,
		&sae.ClientConfig,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to remote server")
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return nil, errors.Wrap(err, "failed to establish a remote session")
	}
	defer session.Close()

	return session.CombinedOutput(action.Command)
}
