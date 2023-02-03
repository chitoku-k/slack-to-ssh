package client

import (
	"context"
	"fmt"

	"github.com/chitoku-k/slack-to-ssh/service"
	"github.com/slack-go/slack"
)

type slackInteractionResponder struct {
	Actions []service.SlackAction
	Client  *slack.Client
}

func NewSlackInteractionResponder(actions []service.SlackAction) service.InteractionResponder {
	return &slackInteractionResponder{
		Actions: actions,
		Client:  slack.New(""),
	}
}

func (sir *slackInteractionResponder) Execute(ctx context.Context, response service.SlackInteractionResponse) error {
	var action *service.SlackAction
	for _, v := range sir.Actions {
		if v.Name == response.ActionName {
			action = &v
			break
		}
	}

	if action == nil {
		return fmt.Errorf("failed to find suitable action: %s", response.ActionName)
	}

	options := []slack.MsgOption{
		slack.MsgOptionResponseURL(response.ResponseURL, slack.ResponseTypeInChannel),
	}

	for _, attachment := range response.Message.Attachments {
		if response.Error != nil {
			attachment.Color = "danger"
			attachment.Text = response.Error.Error()
		} else {
			attachment.Color = "good"
			attachment.Text = action.AttachmentText
		}
		if response.Body != nil {
			attachment.MarkdownIn = []string{"fields"}
			attachment.Fields = []slack.AttachmentField{
				{Value: fmt.Sprintf("```\n%s\n```", response.Body)},
			}
		} else {
			attachment.Fields = nil
		}
		options = append(options, slack.MsgOptionAttachments(attachment))
	}

	_, _, _, err := sir.Client.SendMessageContext(
		ctx,
		response.Message.Channel,
		options...,
	)
	if err != nil {
		return fmt.Errorf("failed to send response: %w", err)
	}
	return nil
}
