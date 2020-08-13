package notifier

//noProvider is the default Provider for notifier.
type noProvider struct{}

func (n noProvider) push(severity, string, string) error {
	return nil
}
