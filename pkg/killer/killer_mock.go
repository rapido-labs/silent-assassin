package killer

import (
	"context"
	"sync"
	"time"

	"github.com/stretchr/testify/mock"
)

type KillerMock struct {
	mock.Mock
}

func (m *KillerMock) EvacuatePodsFromNode(name string, timeout time.Duration, preemption bool) error {
	args := m.Called(name, timeout, preemption)
	return args.Error(0)
}

func (m *KillerMock) Start(ctx context.Context, wg *sync.WaitGroup) {
}
