package shifter

import (
	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/gcloud"
	"github.com/roppenlabs/silent-assassin/pkg/k8s"
	"github.com/roppenlabs/silent-assassin/pkg/killer"
	"github.com/roppenlabs/silent-assassin/pkg/logger"
	"github.com/roppenlabs/silent-assassin/pkg/notifier"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ShifterTestSuit struct {
	suite.Suite
	configMock   *config.ProviderMock
	logger       logger.IZapLogger
	k8sMock      *k8s.K8sClientMock
	gCloudMock   *gcloud.GCloudClientMock
	killer       killer.KillerService
	notifierMock *notifier.NotifierClientMock
}

func (s *ShifterTestSuit) SetupTest() {
	s.configMock = new(config.ProviderMock)
	s.logger = logger.Init(s.configMock)
	s.k8sMock = new(k8s.K8sClientMock)
	s.gCloudMock = new(gcloud.GCloudClientMock)
	s.notifierMock = new(notifier.NotifierClientMock)
	s.notifierMock.On("Info", mock.Anything, mock.Anything)
	s.notifierMock.On("Error", mock.Anything, mock.Anything)
	s.configMock.On("GetString", mock.Anything).Return("debug")
}
