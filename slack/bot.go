package slack

import (
	"github.com/nlopes/slack"
)

type Slack struct {
	slackClient *slack.Client
}

func NewSlack(apiToken string) *Slack {
	return &Slack{
		slackClient: slack.New(apiToken),
	}
}

func (s *Slack) PostMessageWithParameters(channel, text string, params slack.PostMessageParameters) error {
	_, _, _, err := s.slackClient.SendMessage(
		channel,
		slack.MsgOptionText(text, params.EscapeText),
		slack.MsgOptionAttachments(params.Attachments...),
		slack.MsgOptionPostMessageParameters(params),
	)
	return err
}

func (s *Slack) PostMessage(channel, text string) error {
	return s.PostMessageWithParameters(channel, text, slack.NewPostMessageParameters())
}
