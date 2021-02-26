package killer

import (
	"context"
	"sync"
	"time"

	"github.com/stretchr/testify/mock"
)

// KillerMock mocks IKiller interface.
type KillerMock struct {
	mock.Mock
}

// Start is part of IKiller interface.
func (m *KillerMock) Start(ctx context.Context, wg *sync.WaitGroup) {
	m.Called(ctx, wg)
}

// DeletePodsFromNode is part of IKiller interface.
func (m *KillerMock) DeletePodsFromNode(name string, timeout time.Duration, gracePeriodSeconds int) error {
	args := m.Called(name, timeout, gracePeriodSeconds)
	return args.Error(0)
}

// EvictPodsFromNode is part of IKiller interface.
func (m *KillerMock) EvictPodsFromNode(name string, timeout time.Duration, evictDeleteDeadline time.Duration, gracePeriodSeconds int) error {
	args := m.Called(name, timeout, evictDeleteDeadline, gracePeriodSeconds)
	return args.Error(0)
}
