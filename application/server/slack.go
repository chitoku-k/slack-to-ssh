package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
)

func (e *engine) HandleSlackEvent(c *gin.Context) {
	verifier, err := slack.NewSecretsVerifier(c.Request.Header, e.Environment.SlackAppSecret)
	if err != nil {
		c.Status(http.StatusUnauthorized)
		c.Error(fmt.Errorf("failed to verify secret: %w", err))
		return
	}

	var builder strings.Builder
	io.Copy(&builder, io.TeeReader(c.Request.Body, &verifier))

	err = verifier.Ensure()
	if err != nil {
		c.Status(http.StatusUnauthorized)
		c.Error(fmt.Errorf("failed to validate request: %w", err))
		return
	}

	values, err := url.ParseQuery(builder.String())
	if err != nil {
		c.Status(http.StatusBadRequest)
		c.Error(fmt.Errorf("failed to parse request body: %w", err))
		return
	}

	var interaction slack.InteractionCallback
	err = json.Unmarshal([]byte(values.Get("payload")), &interaction)
	if err != nil {
		c.Status(http.StatusBadRequest)
		c.Error(fmt.Errorf("failed to parse interaction: %w", err))
		return
	}

	go func() {
		for _, action := range interaction.ActionCallback.AttachmentActions {
			result, err := e.ActionService.Execute(action.Value)
			if err != nil {
				if result != nil {
					c.Error(fmt.Errorf("failed to execute command: %w", err))

					err = e.InteractionService.Fail(*action, interaction, result, err)
					if err != nil {
						c.Error(fmt.Errorf("failed to respond to interaction: %w", err))
					}
					return
				}

				c.Error(fmt.Errorf("failed to execute action: %w", err))
				return
			}

			err = e.InteractionService.Respond(*action, interaction)
			if err != nil {
				c.Error(fmt.Errorf("failed to respond to interaction: %w", err))
			}
		}
	}()

	c.Status(http.StatusOK)
}
