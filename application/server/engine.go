package server

import (
	"net/http"

	"github.com/chitoku-k/slack-to-ssh/infrastructure/config"
	"github.com/chitoku-k/slack-to-ssh/service"
	"github.com/gin-gonic/gin"
)

type engine struct {
	Environment        config.Environment
	ActionService      service.ActionService
	InteractionService service.InteractionService
}

type Engine interface {
	Start()
}

func NewEngine(
	environment config.Environment,
	action service.ActionService,
	interaction service.InteractionService,
) Engine {
	return &engine{
		Environment:        environment,
		ActionService:      action,
		InteractionService: interaction,
	}
}

func (e *engine) Start() {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/healthz"},
	}))

	router.Any("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	router.POST("/", e.HandleSlackEvent)

	router.Run(":" + e.Environment.Port)
}
