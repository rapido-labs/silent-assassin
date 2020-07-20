package notifier

type NoProvider struct{}

func (n NoProvider) push(severity, string, string) error {
	return nil
}
