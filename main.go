package main

import (
	"fmt"

	"github.com/chitoku-k/slack-to-ssh/application/server"
	"github.com/chitoku-k/slack-to-ssh/infrastructure/client"
	"github.com/chitoku-k/slack-to-ssh/infrastructure/config"
	"github.com/chitoku-k/slack-to-ssh/service"
)

func main() {
	env, err := config.Get()
	if err != nil {
		panic(fmt.Errorf("failed to initialize config: %w", err))
	}

	shellActionExecutor, err := client.NewShellActionExecutor(env)
	if err != nil {
		panic(fmt.Errorf("faled to initialize ssh config: %w", err))
	}

	action := service.NewActionService(shellActionExecutor)
	responder := client.NewSlackInteractionResponder(env)
	interaction := service.NewInteractionService(responder)

	engine := server.NewEngine(env, action, interaction)
	engine.Start()
}
