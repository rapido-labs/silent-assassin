package notifier

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/roppenlabs/silent-assassin/pkg/config"
)

//Slack contains information about slack webhook
type Slack struct {
	url      string
	username string
	channel  string
	iconURL  string
}

//SlackPayload holds the channel and Attachments
type SlackPayload struct {
	Channel     string            `json:"channel"`
	Username    string            `json:"username"`
	IconURL     string            `json:"icon_url"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

//SlackAttachment is sds
type SlackAttachment struct {
	Severity severity     `json:"color"`
	Blocks   []SlackBlock `json:"blocks,omitempty"`
}

//SlackBlock holds the markdown message body.
type SlackBlock struct {
	BlockType string     `json:"type"`
	Text      *SlackText `json:"text,omitempty"`
}

//SlackText asdsad
type SlackText struct {
	TextType string `json:"type"`
	Text     string `json:"text"`
}

//NewSlackClient generates a new Slack client
func NewSlackClient(cp config.IProvider) (Slack, error) {
	var slack Slack
	hookURL := cp.GetString(config.SlackWebhookURL)
	username := cp.GetString(config.SlackUsername)
	channel := cp.GetString(config.SlackChannel)
	iconURL := cp.GetString(config.SlackIconURL)

	_, err := url.ParseRequestURI(hookURL)
	if err != nil {
		return slack, fmt.Errorf("invalid Slack hook URL %s", hookURL)
	}
	if username == "" {
		return slack, errors.New("empty Slack username")
	}

	if channel == "" {
		return slack, errors.New("empty Slack channel")
	}

	return Slack{
		url:      hookURL,
		username: username,
		channel:  fmt.Sprintf("#%s", channel),
		iconURL:  iconURL,
	}, nil
}

//createPayload creates request payload for slack webhook
func (s Slack) createPayload(severity severity, title, details string) SlackPayload {
	titleBlock := SlackBlock{
		BlockType: "section",
		Text: &SlackText{
			TextType: "mrkdwn",
			Text:     fmt.Sprintf("*:bell: %s*", title),
		},
	}
	detailText := &SlackText{
		TextType: "mrkdwn",
		Text:     fmt.Sprintf("```%s```", details),
	}

	detailBlock := SlackBlock{
		BlockType: "section",
		Text:      detailText,
	}

	attachment := SlackAttachment{
		Blocks:   []SlackBlock{titleBlock, detailBlock},
		Severity: severity,
	}
	payload := SlackPayload{
		Channel:     s.channel,
		Username:    s.username,
		IconURL:     s.iconURL,
		Attachments: []SlackAttachment{attachment},
		// Blocks:   []SlackBlock{titleBlock, detailBlock},
	}

	return payload

}

//push sends the notificatio to Slack webhook
func (s Slack) push(severity severity, title, details string) error {

	payload := s.createPayload(severity, title, details)
	err := postMessage(s.url, payload)

	return err
}
