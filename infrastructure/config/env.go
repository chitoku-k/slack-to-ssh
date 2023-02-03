package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/chitoku-k/slack-to-ssh/service"
	"github.com/sirupsen/logrus"
)

const DefaultSSHPort = "22"

type Environment struct {
	SSH SSH

	SlackActions []service.SlackAction

	SlackAppSecret string
	Port           string
	TLSCert        string
	TLSKey         string
}

type SSH struct {
	HostName   string
	Port       string
	Username   string
	KnownHosts []byte
	PrivateKey []byte
}

func Get() (Environment, error) {
	var missing []string
	var env Environment

	var sshKnownHostsPath string
	var sshPrivateKeyPath string

	for k, v := range map[string]*string{
		"SSH_PORT":             &env.SSH.Port,
		"SSH_KNOWN_HOSTS_FILE": &sshKnownHostsPath,
		"TLS_CERT":             &env.TLSCert,
		"TLS_KEY":              &env.TLSKey,
	} {
		*v = os.Getenv(k)
	}

	for k, v := range map[string]*string{
		"SSH_HOSTNAME":         &env.SSH.HostName,
		"SSH_USERNAME":         &env.SSH.Username,
		"SSH_PRIVATE_KEY_FILE": &sshPrivateKeyPath,
		"SLACK_APP_SECRET":     &env.SlackAppSecret,
		"PORT":                 &env.Port,
	} {
		*v = os.Getenv(k)

		if *v == "" {
			missing = append(missing, k)
		}
	}

	for i := 0; ; i++ {
		name := os.Getenv(fmt.Sprintf("SLACK_ACTION_%d_NAME", i))
		if name == "" {
			break
		}

		action := service.SlackAction{Name: name}
		for k, v := range map[string]*string{
			fmt.Sprintf("SLACK_ACTION_%d_COMMAND", i): &action.Command,
		} {
			*v = os.Getenv(k)

			if *v == "" {
				missing = append(missing, k)
			}
		}

		for k, v := range map[string]*string{
			fmt.Sprintf("SLACK_ACTION_%d_ATTACHMENT_TEXT", i): &action.AttachmentText,
		} {
			*v = os.Getenv(k)
		}

		env.SlackActions = append(env.SlackActions, action)
	}

	if len(missing) > 0 {
		return env, errors.New("missing env(s): " + strings.Join(missing, ", "))
	}

	if len(env.SlackActions) == 0 {
		return env, errors.New("missing actions: at least 1 action is required")
	}

	var err error
	if sshKnownHostsPath != "" {
		env.SSH.KnownHosts, err = os.ReadFile(sshKnownHostsPath)
		if err != nil {
			return env, fmt.Errorf("failed to read known hosts %q: %w", sshKnownHostsPath, err)
		}
	} else {
		logrus.Warnln("SSH host key verification is disabled. Consider configuring SSH_KNOWN_HOSTS_FILE.")
	}

	env.SSH.PrivateKey, err = os.ReadFile(sshPrivateKeyPath)
	if err != nil {
		return env, fmt.Errorf("failed to read private key %q: %w", sshPrivateKeyPath, err)
	}

	if env.SSH.Port == "" {
		env.SSH.Port = DefaultSSHPort
	}

	return env, nil
}
