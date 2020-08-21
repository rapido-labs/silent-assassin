package notifier

import (
	"context"
	"fmt"
	"sync"

	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/logger"
)

type severity string

// DANGER AND GOOD are two severity levels for notifications.
const (
	DANGER severity = "#FF0000"
	GOOD   severity = "#006400"
)

type Notification struct {
	Severity severity
	Title    string
	Details  string
}

//NotificationService is a notification engine
type NotificationService struct {
	notificationEvent chan Notification
	provider          Provider
}

//NewNotifier creates a new notifier client
func NewNotificationService(cp config.IProvider, zl logger.IZapLogger) *NotificationService {
	var err error
	var provider Provider
	provider = noProvider{}
	if cp.GetString(config.SlackWebhookURL) != "" {
		provider, err = NewSlackClient(cp)
	}
	if err != nil {
		zl.Error(fmt.Sprintf("Error configuring Slack: %s", err))
		provider = noProvider{}
	}

	return &NotificationService{
		provider:          provider,
		notificationEvent: make(chan Notification),
	}
}

//Start starts the notifier service
func (n NotificationService) Start(ctx context.Context, wg *sync.WaitGroup) {
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
