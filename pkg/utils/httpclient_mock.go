package utils

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Mock.Called(req)
	return args.Get(0).(*http.Response), args.Error(0)
}
