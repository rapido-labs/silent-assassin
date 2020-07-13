package notifier

import (
	"context"
	"fmt"
	"sync"

	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/logger"
)

type severity string

// DANGER, WARNING AND GOOD are two severity levels for notifications.
const (
	DANGER severity = "#FF0000"
	GOOD   severity = "#006400"
)

type Data struct {
	Severity severity
	Title    string
	Details  string
}

//Notifier is a notification engine
type Notifier struct {
	notificationEvent chan Data
	provider          Provider
}

//NewNotifier creates a new notifier client
func NewNotifier(cp config.IProvider, zl logger.IZapLogger) *Notifier {
	var err error
	var provider Provider
	provider = NoProvider{}
	if cp.GetString(config.SlackWebhookURL) != "" {
		provider, err = NewSlackClient(cp)
	}
	if err != nil {
		zl.Error(fmt.Sprintf("Error configuring Slack: %s", err))
		provider = NoProvider{}
	}

	return &Notifier{
		provider:          provider,
		notificationEvent: make(chan Data),
	}
}

//Start starts the notifier service
func (n *Notifier) Start(ctx context.Context, wg *sync.WaitGroup) {
	for {
		select {
		case <-ctx.Done():
			wg.Done()
			return
		case data := <-n.notificationEvent:
			go n.provider.push(data.Severity, data.Title, data.Details)
		}
	}
}

type INotifierClient interface {
	Info(event, details string)
	Error(event, details string)
}

// Provider is a Messaging interface
type Provider interface {
	push(severity severity, title, details string) error
}

func (n *Notifier) publish(severity severity, event, details string) {
	data := Data{
		Severity: severity,
		Title:    event,
		Details:  details,
	}
	n.notificationEvent <- data

}

//Info is for pushing events of level Info, this will print notifications in green color
func (n *Notifier) Info(event, details string) {
	n.publish(GOOD, event, details)
}

//Error is for pushing events of level error, this will print notifications in red color
func (n *Notifier) Error(event, details string) {
	n.publish(DANGER, event, details)
}
