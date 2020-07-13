package notifier

import "github.com/stretchr/testify/mock"

type NotifierMock struct {
	mock.Mock
}

func (m NotifierMock) Info(event, details string) {
}

func (m NotifierMock) Error(event, details string) {
}
