package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/chitoku-k/slack-to-ssh/application/server"
	"github.com/chitoku-k/slack-to-ssh/infrastructure/client"
	"github.com/chitoku-k/slack-to-ssh/infrastructure/config"
	"github.com/chitoku-k/slack-to-ssh/service"
)

var signals = []os.Signal{os.Interrupt}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), signals...)
	defer stop()

	env, err := config.Get()
	if err != nil {
		slog.Error("Failed to initialize config", slog.Any("err", err))
		os.Exit(1)
	}

	shellActionExecutor, err := client.NewShellActionExecutor(env.SlackActions, env.SSH.HostName, env.SSH.Port, env.SSH.Username, env.SSH.KnownHosts, env.SSH.PrivateKey)
	if err != nil {
		slog.Error("Failed to initialize ssh config", slog.Any("err", err))
		os.Exit(1)
	}

	action := service.NewActionService(shellActionExecutor)
	responder := client.NewSlackInteractionResponder(env.SlackActions)
	interaction := service.NewInteractionService(responder)

	engine := server.NewEngine(env.Port, env.TLSCert, env.TLSKey, env.SlackAppSecret, action, interaction)
	err = engine.Start(ctx)
	if err != nil {
		slog.Error("Failed to start web server", slog.Any("err", err))
		os.Exit(1)
	}
}
