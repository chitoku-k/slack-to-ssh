package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chitoku-k/slack-to-ssh/service"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
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
	KnownHosts []ssh.PublicKey
	PrivateKey ssh.Signer
}

func Get() (Environment, error) {
	var missing []string
	var env Environment

	var sshKnownHostsPath string
	var sshPrivateKeyPath, sshPrivateKeyPassphrasePath string

	for k, v := range map[string]*string{
		"SSH_PORT":                        &env.SSH.Port,
		"SSH_KNOWN_HOSTS_FILE":            &sshKnownHostsPath,
		"SSH_PRIVATE_KEY_PASSPHRASE_FILE": &sshPrivateKeyPassphrasePath,
		"TLS_CERT":                        &env.TLSCert,
		"TLS_KEY":                         &env.TLSKey,
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

	if sshKnownHostsPath != "" {
		knownHosts, err := os.ReadFile(sshKnownHostsPath)
		if err != nil {
			return env, fmt.Errorf("failed to read known hosts %q: %w", sshKnownHostsPath, err)
		}

		for knownHosts != nil {
			publicKey, rest, err := parseSSHKnownHosts(knownHosts, env.SSH.HostName)
			if err != nil {
				return env, fmt.Errorf("failed to parse known hosts: %w", err)
			}
			if publicKey != nil {
				env.SSH.KnownHosts = append(env.SSH.KnownHosts, publicKey)
			}
			knownHosts = rest
		}
	} else {
		logrus.Warnln("SSH host key verification is disabled. Consider configuring SSH_KNOWN_HOSTS_FILE.")
	}

	sshPrivateKey, err := os.ReadFile(sshPrivateKeyPath)
	if err != nil {
		return env, fmt.Errorf("failed to read private key %q: %w", sshPrivateKeyPath, err)
	}

	if sshPrivateKeyPassphrasePath != "" {
		sshPrivateKeyPassphrase, err := os.ReadFile(sshPrivateKeyPassphrasePath)
		if err != nil {
			return env, fmt.Errorf("failed to read passphrase %q: %w", sshPrivateKeyPassphrasePath, err)
		}
		env.SSH.PrivateKey, err = ssh.ParsePrivateKeyWithPassphrase(sshPrivateKey, sshPrivateKeyPassphrase)
	} else {
		env.SSH.PrivateKey, err = ssh.ParsePrivateKey(sshPrivateKey)
	}

	if err != nil {
		return env, fmt.Errorf("faile to parse private key %q: %w", sshPrivateKeyPath, err)
	}

	if env.SSH.Port == "" {
		env.SSH.Port = DefaultSSHPort
	}

	return env, nil
}

func parseSSHKnownHosts(knownHosts []byte, hostname string) (publicKey ssh.PublicKey, rest []byte, err error) {
	marker, hosts, publicKey, _, rest, err := ssh.ParseKnownHosts(knownHosts)
	if err == io.EOF {
		return nil, nil, nil
	}
	if err != nil {
		return nil, rest, err
	}
	if marker == "revoked" {
		return nil, rest, err
	}
	for _, h := range hosts {
		if h == hostname {
			return publicKey, rest, err
		}
	}
	return nil, rest, err
}
