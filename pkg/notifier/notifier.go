package notifier

type severity string

// DANGER, WARNING AND GOOD are various severity levels for notifications.
const (
	DANGER severity = "#FF0000"
	GOOD   severity = "#006400"
)

//Notifier is a notification engine
type Notifier struct {
	messageClient MessageClient
}

type INotifier interface {
	Info(event, details string) error
	Error(event, details string) error
}

// MessageClient is a Messaging interface
type MessageClient interface {
	push(severity severity, title, details string) error
}

//NewNotifier creates a new notifier client
func NewNotifier(client MessageClient) Notifier {
	return Notifier{messageClient: client}
}

func (n Notifier) Info(event, details string) error {
	return n.messageClient.push(GOOD, event, details)
}

func (n Notifier) Error(event, details string) error {
	return n.messageClient.push(DANGER, event, details)
}
