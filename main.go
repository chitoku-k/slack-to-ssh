package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/chitoku-k/slack-to-ssh/application/server"
	"github.com/chitoku-k/slack-to-ssh/infrastructure/client"
	"github.com/chitoku-k/slack-to-ssh/infrastructure/config"
	"github.com/chitoku-k/slack-to-ssh/service"
	"github.com/sirupsen/logrus"
)

var signals = []os.Signal{os.Interrupt}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), signals...)
	defer stop()

	env, err := config.Get()
	if err != nil {
		logrus.Fatalf("Failed to initialize config: %v", err)
	}

	shellActionExecutor, err := client.NewShellActionExecutor(env.SlackActions, env.SSH.HostName, env.SSH.Port, env.SSH.Username, env.SSH.KnownHosts, env.SSH.PrivateKey)
	if err != nil {
		logrus.Fatalf("Failed to initialize ssh config: %v", err)
	}

	action := service.NewActionService(shellActionExecutor)
	responder := client.NewSlackInteractionResponder(env.SlackActions)
	interaction := service.NewInteractionService(responder)

	engine := server.NewEngine(env.Port, env.TLSCert, env.TLSKey, env.SlackAppSecret, action, interaction)
	err = engine.Start(ctx)
	if err != nil {
		logrus.Fatalf("Failed to start web server: %v", err)
	}
}
