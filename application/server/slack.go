package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
)

func (e *engine) HandleSlackEvent(ctx context.Context, c *gin.Context) {
	verifier, err := slack.NewSecretsVerifier(c.Request.Header, e.SlackAppSecret)
	if err != nil {
		c.Status(http.StatusUnauthorized)
		c.Error(fmt.Errorf("failed to verify secret: %w", err))
		return
	}

	var builder strings.Builder
	_, err = io.Copy(&builder, io.TeeReader(c.Request.Body, &verifier))
	if err != nil {
		c.Status(http.StatusBadRequest)
		c.Error(fmt.Errorf("failed to read body: %w", err))
		return
	}

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
			result, err := e.ActionService.Execute(ctx, action.Value)
			if err != nil {
				if result != nil {
					c.Error(fmt.Errorf("failed to execute command: %w", err))

					err = e.InteractionService.Fail(ctx, *action, interaction, result, err)
					if err != nil {
						c.Error(fmt.Errorf("failed to respond to interaction: %w", err))
					}
					return
				}

				c.Error(fmt.Errorf("failed to execute action: %w", err))
				return
			}

			err = e.InteractionService.Respond(ctx, *action, interaction)
			if err != nil {
				c.Error(fmt.Errorf("failed to respond to interaction: %w", err))
			}
		}
	}()

	c.Status(http.StatusOK)
}
