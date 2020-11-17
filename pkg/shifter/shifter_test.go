package shifter

import (
	"testing"

	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/gcloud"
	"github.com/roppenlabs/silent-assassin/pkg/k8s"
	"github.com/roppenlabs/silent-assassin/pkg/killer"
	"github.com/roppenlabs/silent-assassin/pkg/logger"
	"github.com/roppenlabs/silent-assassin/pkg/notifier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (st *ShifterTestSuit) SetupTest() {
	st.configMock = new(config.ProviderMock)
	st.k8sMock = new(k8s.K8sClientMock)
	st.gCloudMock = new(gcloud.GCloudClientMock)
	st.notifierMock = new(notifier.NotifierClientMock)
	st.notifierMock.On("Info", mock.Anything, mock.Anything)
	st.notifierMock.On("Error", mock.Anything, mock.Anything)
	st.configMock.On("GetString", mock.Anything).Return("debug")
	st.logger = logger.Init(st.configMock)
}

func (st *ShifterTestSuit) TestShouldReturnRightNodePoolSize() {
	nodesEquallyDistributed := []v1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "node-1",
				Labels: map[string]string{"failure-domain.beta.kubernetes.io/zone": "asia-south1-a"},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "node-2",
				Labels: map[string]string{"failure-domain.beta.kubernetes.io/zone": "asia-south1-b"},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "node-3",
				Labels: map[string]string{"failure-domain.beta.kubernetes.io/zone": "asia-south1-c"},
			},
		},
	}

	nodeInEquallyDistributed := []v1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "node-1",
				Labels: map[string]string{"failure-domain.beta.kubernetes.io/zone": "asia-south1-a"},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "node-2",
				Labels: map[string]string{"failure-domain.beta.kubernetes.io/zone": "asia-south1-a"},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "node-3",
				Labels: map[string]string{"failure-domain.beta.kubernetes.io/zone": "asia-south1-a"},
			},
		},
	}

	st.k8sMock.On("GetNodes", "cloud.google.com/gke-nodepool=nodepool-1").Return(&v1.NodeList{Items: nodesEquallyDistributed}, nil)
	st.k8sMock.On("GetNodes", "cloud.google.com/gke-nodepool=nodepool-2").Return(&v1.NodeList{Items: nodeInEquallyDistributed}, nil)

	kl := killer.NewKillerService(st.configMock, st.logger, st.k8sMock, st.gCloudMock, st.notifierMock)
	ss := NewShifterService(st.configMock, st.logger, st.k8sMock, st.gCloudMock, st.notifierMock, kl)

	size, err := ss.getNodePoolSize("cloud.google.com/gke-nodepool=nodepool-1")

	assert.Equal(st.T(), int64(1), size, "Expected 1 as size")
	assert.Nil(st.T(), err)

	size, err = ss.getNodePoolSize("cloud.google.com/gke-nodepool=nodepool-2")

	assert.Nil(st.T(), err)
	assert.Equal(st.T(), int64(3), size, "Expected 3 as size")

}

func TestKillerTestSuite(t *testing.T) {
	suite.Run(t, new(ShifterTestSuit))
}
