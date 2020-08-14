package gcloud

import "github.com/stretchr/testify/mock"

type GCloudClientMock struct {
	mock.Mock
}

func (m *GCloudClientMock) DeleteInstance(projectID, zone, name string) error {
	args := m.Called(projectID, name)
	return args.Error(0)
}
