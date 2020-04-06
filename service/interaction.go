package service

import (
	"github.com/slack-go/slack"
)

type interactionService struct {
	InteractionResponder InteractionResponder
}

type InteractionService interface {
	Respond(action slack.AttachmentAction, interaction slack.InteractionCallback) error
	Fail(action slack.AttachmentAction, interaction slack.InteractionCallback, body []byte, err error) error
}

type InteractionResponder interface {
	Execute(response SlackInteractionResponse) error
}

func NewInteractionService(responder InteractionResponder) InteractionService {
	return &interactionService{
		InteractionResponder: responder,
	}
}

func (is *interactionService) Respond(action slack.AttachmentAction, interaction slack.InteractionCallback) error {
	return is.InteractionResponder.Execute(SlackInteractionResponse{
		ActionName:  action.Value,
		ResponseURL: interaction.ResponseURL,
		Message:     interaction.OriginalMessage,
	})
}

func (is *interactionService) Fail(action slack.AttachmentAction, interaction slack.InteractionCallback, body []byte, err error) error {
	return is.InteractionResponder.Execute(SlackInteractionResponse{
		ActionName:  action.Value,
		ResponseURL: interaction.ResponseURL,
		Message:     interaction.OriginalMessage,
		Body:        body,
		Error:       err,
	})
}
