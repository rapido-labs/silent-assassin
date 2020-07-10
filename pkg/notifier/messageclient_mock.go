package notifier

import "github.com/stretchr/testify/mock"

//MessageClientMock mocks the MessageClient struct
type MessageClientMock struct {
	mock.Mock
}

//Push mocks message push to destination
func (m *MessageClientMock) push(severity severity, title, details string) error {
	args := m.Called(severity, title, details)
	return args.Error(0)
}
