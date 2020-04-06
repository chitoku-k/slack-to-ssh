package service

import "github.com/slack-go/slack"

type SlackAction struct {
	Name           string
	AttachmentText string
	Command        string
}

type SlackInteractionResponse struct {
	ActionName  string
	ResponseURL string
	Message     slack.Message
	Body        []byte
	Error       error
}
