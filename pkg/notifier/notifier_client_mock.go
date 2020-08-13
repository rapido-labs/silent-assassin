package notifier

import "github.com/stretchr/testify/mock"

type NotifierClientMock struct {
	mock.Mock
}

func (m NotifierClientMock) Info(event, details string) {
}

func (m NotifierClientMock) Error(event, details string) {
}
