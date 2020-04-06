package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/chitoku-k/slack-to-ssh/service"
)

const DefaultSSHPort = "22"

type Environment struct {
	SSH SSH

	SlackActions []service.SlackAction

	SlackAppSecret string
	Port           string
}

type SSH struct {
	HostName   string
	Port       string
	Username   string
	PrivateKey string
}

func Get() (Environment, error) {
	var missing []string
	var env Environment

	for k, v := range map[string]*string{
		"SSH_HOSTNAME":     &env.SSH.HostName,
		"SSH_PORT":         &env.SSH.Port,
		"SSH_USERNAME":     &env.SSH.Username,
		"SSH_PRIVATE_KEY":  &env.SSH.PrivateKey,
		"SLACK_APP_SECRET": &env.SlackAppSecret,
		"PORT":             &env.Port,
	} {
		*v = os.Getenv(k)

		if k == "SSH_PORT" && *v == "" {
			*v = DefaultSSHPort
			continue
		}

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
			fmt.Sprintf("SLACK_ACTION_%d_COMMAND", i):         &action.Command,
			fmt.Sprintf("SLACK_ACTION_%d_ATTACHMENT_TEXT", i): &action.AttachmentText,
		} {
			*v = os.Getenv(k)

			if *v == "" {
				missing = append(missing, k)
			}
		}

		env.SlackActions = append(env.SlackActions, action)
	}

	if len(missing) > 0 {
		return env, errors.New("missing env(s): " + strings.Join(missing, ", "))
	}

	if len(env.SlackActions) == 0 {
		return env, errors.New("missing actions: at least 1 action is required")
	}

	return env, nil
}
