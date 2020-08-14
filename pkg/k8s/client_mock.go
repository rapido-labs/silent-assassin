package k8s

import (
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
)

type K8sClientMock struct {
	mock.Mock
}

func (m *K8sClientMock) GetNodes(labelSelector string) *v1.NodeList {
	args := m.Called(labelSelector)
	return args.Get(0).(*v1.NodeList)
}

func (m *K8sClientMock) GetNode(name string) (v1.Node, error) {
	args := m.Called(name)
	return args.Get(0).(v1.Node), args.Error(0)
}

func (m *K8sClientMock) UpdateNode(node v1.Node) error {
	args := m.Called(node)
	return args.Error(0)
}

func (m *K8sClientMock) DeleteNode(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *K8sClientMock) DeletePod(name, namespace string) error {
	args := m.Called(name, namespace)
	return args.Error(0)
}

func (m *K8sClientMock) GetPodsInNode(name string) ([]v1.Pod, error) {
	args := m.Called(name)
	return args.Get(0).([]v1.Pod), args.Error(1)
}
