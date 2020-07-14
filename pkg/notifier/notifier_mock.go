package notifier

import "github.com/stretchr/testify/mock"

type NotifierMock struct {
	mock.Mock
}

func (m NotifierMock) Info(event, details string) error {
	args := m.Called(event, details)
	return args.Error(0)
}

func (m NotifierMock) Error(event, details string) error {
	args := m.Called(event, details)
	return args.Error(0)
}
