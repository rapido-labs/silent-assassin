package killer

import (
	"context"
	"sync"

	"github.com/stretchr/testify/mock"
)

type KillerMock struct {
	mock.Mock
}

func (m *KillerMock) EvacuatePodsFromNode(name string, timeout uint32, preemption bool) error {
	args := m.Called(name, timeout, preemption)
	return args.Error(0)
}

func (m *KillerMock) Start(ctx context.Context, wg *sync.WaitGroup) {
}
