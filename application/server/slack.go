package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/slack-go/slack"
)

func (e *engine) HandleSlackEvent(c *gin.Context) {
	verifier, err := slack.NewSecretsVerifier(c.Request.Header, e.Environment.SlackAppSecret)
	if err != nil {
		c.Status(http.StatusUnauthorized)
		c.Error(errors.Wrap(err, "failed to verify secret"))
		return
	}

	var builder strings.Builder
	io.Copy(&builder, io.TeeReader(c.Request.Body, &verifier))

	err = verifier.Ensure()
	if err != nil {
		c.Status(http.StatusUnauthorized)
		c.Error(errors.Wrap(err, "failed to validate request"))
		return
	}

	values, err := url.ParseQuery(builder.String())
	if err != nil {
		c.Status(http.StatusBadRequest)
		c.Error(errors.Wrap(err, "failed to parse request body"))
		return
	}

	var interaction slack.InteractionCallback
	err = json.Unmarshal([]byte(values.Get("payload")), &interaction)
	if err != nil {
		c.Status(http.StatusBadRequest)
		c.Error(errors.Wrap(err, "failed to parse interaction"))
		return
	}

	go func() {
		for _, action := range interaction.ActionCallback.AttachmentActions {
			result, err := e.ActionService.Execute(action.Value)
			if err != nil {
				if result != nil {
					c.Error(errors.Wrap(err, "failed to execute command"))

					err = e.InteractionService.Fail(*action, interaction, result, err)
					if err != nil {
						c.Error(errors.Wrap(err, "failed to respond to interaction"))
					}
					return
				}

				c.Error(errors.Wrap(err, "failed to execute action"))
				return
			}

			err = e.InteractionService.Respond(*action, interaction)
			if err != nil {
				c.Error(errors.Wrap(err, "failed to respond to interaction"))
			}
		}
	}()

	c.Status(http.StatusOK)
}
