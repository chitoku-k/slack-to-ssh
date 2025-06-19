package server

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
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
		slog.Error("Failed to verify secret", slog.Any("err", err))
		return
	}

	var builder strings.Builder
	_, err = io.Copy(&builder, io.TeeReader(c.Request.Body, &verifier))
	if err != nil {
		c.Status(http.StatusBadRequest)
		slog.Error("Failed to read body", slog.Any("err", err))
		return
	}

	err = verifier.Ensure()
	if err != nil {
		c.Status(http.StatusUnauthorized)
		slog.Error("Failed to validate request", slog.Any("err", err))
		return
	}

	values, err := url.ParseQuery(builder.String())
	if err != nil {
		c.Status(http.StatusBadRequest)
		slog.Error("Failed to parse request body", slog.Any("err", err))
		return
	}

	var interaction slack.InteractionCallback
	err = json.Unmarshal([]byte(values.Get("payload")), &interaction)
	if err != nil {
		c.Status(http.StatusBadRequest)
		slog.Error("Failed to parse interaction", slog.Any("err", err))
		return
	}

	go func() {
		for _, action := range interaction.ActionCallback.AttachmentActions {
			result, err := e.ActionService.Execute(ctx, action.Value)
			if err != nil {
				if result != nil {
					slog.Error("Failed to execute command", slog.Any("err", err))

					err = e.InteractionService.Fail(ctx, *action, interaction, result, err)
					if err != nil {
						slog.Error("Failed to respond to interaction", slog.Any("err", err))
					}
					return
				}

				slog.Error("Failed to execute action", slog.Any("err", err))
				return
			}

			err = e.InteractionService.Respond(ctx, *action, interaction)
			if err != nil {
				slog.Error("Failed to respond to interaction", slog.Any("err", err))
			}
		}
	}()

	c.Status(http.StatusOK)
}
