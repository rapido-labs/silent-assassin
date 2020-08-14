package notifier

type INotifierClient interface {
	Info(event, details string)
	Error(event, details string)
}

//publish publishes the notifications to the notificationEventchannel od Notifier struct.
func (n Notifier) publish(severity severity, event, details string) {
	data := Notification{
		Severity: severity,
		Title:    event,
		Details:  details,
	}
	n.notificationEvent <- data

}

//Info is for pushing events of level Info, this will print notifications in green color
func (n Notifier) Info(event, details string) {
	n.publish(GOOD, event, details)
}

//Error is for pushing events of level error, this will print notifications in red color
func (n Notifier) Error(event, details string) {
	n.publish(DANGER, event, details)
}
