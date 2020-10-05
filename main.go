package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/chitoku-k/slack-to-ssh/application/server"
	"github.com/chitoku-k/slack-to-ssh/infrastructure/client"
	"github.com/chitoku-k/slack-to-ssh/infrastructure/config"
	"github.com/chitoku-k/slack-to-ssh/service"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		defer close(sig)
		<-sig
		cancel()
	}()

	env, err := config.Get()
	if err != nil {
		logrus.Fatalf("Failed to initialize config: %v", err)
	}

	shellActionExecutor, err := client.NewShellActionExecutor(env)
	if err != nil {
		logrus.Fatalf("Failed to initialize ssh config: %v", err)
	}

	action := service.NewActionService(shellActionExecutor)
	responder := client.NewSlackInteractionResponder(env)
	interaction := service.NewInteractionService(responder)

	engine := server.NewEngine(env.Port, env.SlackAppSecret, action, interaction)
	err = engine.Start(ctx)
	if err != nil {
		logrus.Fatalf("Failed to start web server: %v", err)
	}
}
