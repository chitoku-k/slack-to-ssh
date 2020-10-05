package service

import (
	"context"

	"github.com/slack-go/slack"
)

type interactionService struct {
	InteractionResponder InteractionResponder
}

type InteractionService interface {
	Respond(ctx context.Context, action slack.AttachmentAction, interaction slack.InteractionCallback) error
	Fail(ctx context.Context, action slack.AttachmentAction, interaction slack.InteractionCallback, body []byte, err error) error
}

type InteractionResponder interface {
	Execute(ctx context.Context, response SlackInteractionResponse) error
}

func NewInteractionService(responder InteractionResponder) InteractionService {
	return &interactionService{
		InteractionResponder: responder,
	}
}

func (is *interactionService) Respond(ctx context.Context, action slack.AttachmentAction, interaction slack.InteractionCallback) error {
	return is.InteractionResponder.Execute(ctx, SlackInteractionResponse{
		ActionName:  action.Value,
		ResponseURL: interaction.ResponseURL,
		Message:     interaction.OriginalMessage,
	})
}

func (is *interactionService) Fail(ctx context.Context, action slack.AttachmentAction, interaction slack.InteractionCallback, body []byte, err error) error {
	return is.InteractionResponder.Execute(ctx, SlackInteractionResponse{
		ActionName:  action.Value,
		ResponseURL: interaction.ResponseURL,
		Message:     interaction.OriginalMessage,
		Body:        body,
		Error:       err,
	})
}
