package notifier

// Provider is a Messaging interface.
// Currently Slack struct implements this.
type Provider interface {
	push(severity severity, title, details string) error
}
