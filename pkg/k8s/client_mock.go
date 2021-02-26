package k8s

import (
	"time"

	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
)

type K8sClientMock struct {
	mock.Mock
}

func (m *K8sClientMock) GetNodes(labelSelector string) (*v1.NodeList, error) {
	args := m.Called(labelSelector)
	return args.Get(0).(*v1.NodeList), args.Error(1)
}

func (m *K8sClientMock) GetNode(name string) (v1.Node, error) {
	args := m.Called(name)
	return args.Get(0).(v1.Node), args.Error(1)
}

func (m *K8sClientMock) UpdateNode(node v1.Node) error {
	args := m.Called(node)
	return args.Error(0)
}

func (m *K8sClientMock) DeleteNode(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *K8sClientMock) GetPodsInNode(name string) ([]v1.Pod, error) {
	args := m.Called(name)
	return args.Get(0).([]v1.Pod), args.Error(1)
}

func (m *K8sClientMock) DrainNode(name string, useEvict bool, timeout time.Duration, gracePeriodSeconds int) error {
	args := m.Called(name, useEvict, timeout, gracePeriodSeconds)
	return args.Error(0)
}
