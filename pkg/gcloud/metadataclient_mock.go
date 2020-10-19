package gcloud

import (
	"github.com/stretchr/testify/mock"
)

type MockMClient struct {
	mock.Mock
}

func (m *MockMClient) InstanceName() (string, error) {
	args := m.Called()
	return args.Get(0).(string), args.Error(1)
}

func (m *MockMClient) Subscribe(suffix string, fn func(v string, ok bool) error) error {
	args := m.Called(suffix, fn)
	return args.Error(1)
}
