package main

import (
	"github.com/chitoku-k/slack-to-ssh/application/server"
	"github.com/chitoku-k/slack-to-ssh/infrastructure/client"
	"github.com/chitoku-k/slack-to-ssh/infrastructure/config"
	"github.com/chitoku-k/slack-to-ssh/service"
	"github.com/pkg/errors"
)

func main() {
	env, err := config.Get()
	if err != nil {
		panic(errors.Wrap(err, "failed to initialize config"))
	}

	shellActionExecutor, err := client.NewShellActionExecutor(env)
	if err != nil {
		panic(errors.Wrap(err, "faled to initialize ssh config"))
	}

	action := service.NewActionService(shellActionExecutor)
	responder := client.NewSlackInteractionResponder(env)
	interaction := service.NewInteractionService(responder)

	engine := server.NewEngine(env, action, interaction)
	engine.Start()
}
