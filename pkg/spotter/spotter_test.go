package spotter

import (
	"testing"
	"time"

	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/k8s"
	"github.com/roppenlabs/silent-assassin/pkg/logger"
	"github.com/roppenlabs/silent-assassin/pkg/notifier"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
)

type SpotterTestSuite struct {
	suite.Suite
	k8sMock      *k8s.K8sClientMock
	configMock   *config.ProviderMock
	logger       logger.IZapLogger
	notifierMock *notifier.NotifierMock
}

func (s *SpotterTestSuite) SetupTest() {
	s.configMock = new(config.ProviderMock)
	s.k8sMock = new(k8s.K8sClientMock)
	s.notifierMock = new(notifier.NotifierMock)
	s.configMock.On("GetString", mock.Anything).Return("debug")
	s.configMock.On("GetInt", "spotter.poll_interval_ms").Return(10)
	s.configMock.On("GetStringSlice", "spotter.label_selectors").Return([]string{"cloud.google.com/gke-preemptible=true,label2=test"})

	s.logger = logger.Init(s.configMock)
}

func (suite *SpotterTestSuite) TestShouldFetchNodesWithLabels() {

	suite.configMock.On("GetStringSlice", config.SpotterWhiteListIntervalHours).Return([]string{"00:00-06:00", "12:00-14:00"})
	suite.k8sMock.On("GetNodes", []string{"cloud.google.com/gke-preemptible=true,label2=test"}).Return(&v1.NodeList{})

	ss := NewSpotterService(suite.configMock, suite.logger, suite.k8sMock, suite.notifierMock)
	ss.initWhitelist()

	ss.spot()

	suite.k8sMock.AssertExpectations(suite.T())
}

func (suite *SpotterTestSuite) TestShouldAnnotateIfAbsent() {

	nodeAlreadyAnnotated := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "Node-1",
			Annotations: map[string]string{"silent-assassin/expiry-time": time.Now().String()}}}
	suite.configMock.On("GetStringSlice", config.SpotterWhiteListIntervalHours).Return([]string{"00:00-06:00", "12:00-14:00"})
	nodeToBeAnnotated := v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "Node-2"}}

	nodeList := v1.NodeList{
		Items: []v1.Node{nodeAlreadyAnnotated, nodeToBeAnnotated},
	}
	suite.k8sMock.On("GetNodes", mock.Anything).Return(&nodeList)
	suite.k8sMock.On("AnnotateNode", mock.MatchedBy(func(input v1.Node) bool {

		_, found := input.ObjectMeta.Annotations["silent-assassin/expiry-time"]
		if !found {
			return false
		}
		assert.Equal(suite.T(), "Node-2", input.ObjectMeta.Name, "Node name is not matching")
		return true

	})).Return(nil)

	suite.notifierMock.On("Info", "ANNOTATE", mock.Anything).Return(nil)
	ss := NewSpotterService(suite.configMock, suite.logger, suite.k8sMock, suite.notifierMock)
	ss.initWhitelist()
	ss.spot()

	suite.k8sMock.AssertExpectations(suite.T())
}

func TestSpotterTestSuite(t *testing.T) {
	suite.Run(t, new(SpotterTestSuite))
}
