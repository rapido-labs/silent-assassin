package gcloud

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
	compute "google.golang.org/api/compute/v1"
	container "google.golang.org/api/container/v1"
)

type GCloudClientMock struct {
	mock.Mock
}

func (m *GCloudClientMock) DeleteInstance(zone, name string) error {
	args := m.Called(zone, name)
	return args.Error(0)
}

func (m *GCloudClientMock) GetInstance(project, zone, name string) (*compute.Instance, error) {
	args := m.Called(project, zone, name)
	return args.Get(0).(*compute.Instance), args.Error(1)
}

func (m *GCloudClientMock) ListNodePools() ([]*container.NodePool, error) {
	args := m.Called()
	return args.Get(0).([]*container.NodePool), args.Error(1)
}

func (m *GCloudClientMock) GetNodePool(npName string) (*container.NodePool, error) {
	args := m.Called(npName)
	return args.Get(0).(*container.NodePool), args.Error(1)
}

func (m *GCloudClientMock) waitForOperation(ctx context.Context, operationID string) error {
	args := m.Called(ctx, operationID)
	return args.Error(0)
}

func (m *GCloudClientMock) SetNodePoolSize(npName string, size int64, timeout time.Duration) error {
	args := m.Called(npName, size, timeout)
	return args.Error(0)
}

func (m *GCloudClientMock) GetNumberOfZones() int {
	args := m.Called()
	return args.Get(0).(int)
}
