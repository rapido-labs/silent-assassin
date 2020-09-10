package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/utils"
)

//Slack contains information about slack webhook
type Slack struct {
	url            string
	username       string
	channel        string
	iconURL        string
	messageTimeout uint32
	httpClient     utils.IHTTPClient
}

//SlackPayload holds the channel and Attachments
type slackPayload struct {
	Channel     string            `json:"channel"`
	Username    string            `json:"username"`
	IconURL     string            `json:"icon_url"`
	Attachments []slackAttachment `json:"attachments,omitempty"`
}

//SlackAttachment is sds
type slackAttachment struct {
	Severity severity     `json:"color"`
	Blocks   []slackBlock `json:"blocks,omitempty"`
}

//SlackBlock holds the markdown message body.
type slackBlock struct {
	BlockType string     `json:"type"`
	Text      *slackText `json:"text,omitempty"`
}

//SlackText asdsad
type slackText struct {
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
	messageTimeout := cp.GetUint32(config.SlackTimeoutMs)

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

	if messageTimeout == 0 {
		messageTimeout = 2000
	}

	httpClient := http.DefaultClient

	return Slack{
		url:            hookURL,
		username:       username,
		channel:        fmt.Sprintf("#%s", channel),
		iconURL:        iconURL,
		httpClient:     httpClient,
		messageTimeout: messageTimeout,
	}, nil
}

//createPayload creates request payload for slack webhook
func (s Slack) createPayload(severity severity, title, details string) slackPayload {
	titleBlock := slackBlock{
		BlockType: "section",
		Text: &slackText{
			TextType: "mrkdwn",
			Text:     fmt.Sprintf("*:bell: %s*", title),
		},
	}
	detailText := &slackText{
		TextType: "mrkdwn",
		Text:     fmt.Sprintf("```%s```", details),
	}

	detailBlock := slackBlock{
		BlockType: "section",
		Text:      detailText,
	}

	attachment := slackAttachment{
		Blocks:   []slackBlock{titleBlock, detailBlock},
		Severity: severity,
	}
	payload := slackPayload{
		Channel:     s.channel,
		Username:    s.username,
		IconURL:     s.iconURL,
		Attachments: []slackAttachment{attachment},
	}

	return payload

}

//postMessage posts request message to the given URL
func (s Slack) postMessage(address string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshalling notification payload failed: %w", err)
	}

	b := bytes.NewBuffer(data)

	req, err := http.NewRequest(http.MethodPost, address, b)
	if err != nil {
		return fmt.Errorf("http NewRequest failed: %w", err)
	}
	req.Header.Set("Content-type", "application/json")

	ctx, cancel := context.WithTimeout(req.Context(), time.Duration(s.messageTimeout)*time.Millisecond)
	defer cancel()

	res, err := s.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("sending notification failed: %w", err)
	}

	defer res.Body.Close()
	statusCode := res.StatusCode
	if statusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(res.Body)
		return fmt.Errorf("Sending notification failed: %s", string(body))
	}

	return nil
}

//push implements Provider interface. Sends the notification to Slack webhook.
func (s Slack) push(severity severity, title, details string) error {

	payload := s.createPayload(severity, title, details)
	err := s.postMessage(s.url, payload)

	return err
}
