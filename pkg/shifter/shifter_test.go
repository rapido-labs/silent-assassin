package shifter

import (
	"testing"
	"time"

	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/gcloud"
	"github.com/roppenlabs/silent-assassin/pkg/k8s"
	"github.com/roppenlabs/silent-assassin/pkg/killer"
	"github.com/roppenlabs/silent-assassin/pkg/logger"
	"github.com/roppenlabs/silent-assassin/pkg/notifier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	container "google.golang.org/api/container/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ShifterTestSuit struct {
	suite.Suite
	configMock   *config.ProviderMock
	logger       logger.IZapLogger
	k8sMock      *k8s.K8sClientMock
	gCloudMock   *gcloud.GCloudClientMock
	killerMock   *killer.KillerMock
	notifierMock *notifier.NotifierClientMock
}

func (st *ShifterTestSuit) SetupTest() {
	st.configMock = new(config.ProviderMock)
	st.k8sMock = new(k8s.K8sClientMock)
	st.gCloudMock = new(gcloud.GCloudClientMock)
	st.killerMock = new(killer.KillerMock)
	st.notifierMock = new(notifier.NotifierClientMock)
	st.notifierMock.On("Info", mock.Anything, mock.Anything)
	st.notifierMock.On("Error", mock.Anything, mock.Anything)
	st.configMock.On("GetString", mock.Anything).Return("debug")
	st.logger = logger.Init(st.configMock)
	nodePools := []*container.NodePool{
		{
			Name: "services-p-1",
			Config: &container.NodeConfig{
				Labels: map[string]string{
					"component":   "services",
					"criticality": "1",
				},
				MachineType: "e2-standard-2",
				Preemptible: true,
			},
			Autoscaling: &container.NodePoolAutoscaling{
				MinNodeCount: 0,
			},
		},
		{
			Name: "services-np-1",
			Config: &container.NodeConfig{
				Labels: map[string]string{
					"component":   "services",
					"criticality": "1",
				},
				MachineType: "e2-standard-2",
				Preemptible: false,
			},
			Autoscaling: &container.NodePoolAutoscaling{
				MinNodeCount: 1,
			},
		},
		{
			Name: "services-p-5",
			Config: &container.NodeConfig{
				Labels: map[string]string{
					"component":   "operations",
					"criticality": "1",
				},
				MachineType: "e2-standard-2",
				Preemptible: true,
			},
			Autoscaling: &container.NodePoolAutoscaling{
				MinNodeCount: 0,
			},
		},
		{
			Name: "services-2",
			Config: &container.NodeConfig{
				Labels: map[string]string{
					"component":   "databases",
					"criticality": "1",
				},
				MachineType: "e2-standard-2",
				Preemptible: false,
			},
			Autoscaling: &container.NodePoolAutoscaling{
				MinNodeCount: 0,
			},
		},
	}
	st.gCloudMock.On("ListNodePools").Return(nodePools, nil)
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

	ss := NewShifterService(st.configMock, st.logger, st.k8sMock, st.gCloudMock, st.notifierMock, st.killerMock)

	size, err := ss.getNodePoolSize("cloud.google.com/gke-nodepool=nodepool-1")

	assert.Equal(st.T(), int64(1), size, "Expected 1 as size")
	assert.Nil(st.T(), err)

	size, err = ss.getNodePoolSize("cloud.google.com/gke-nodepool=nodepool-2")

	assert.Nil(st.T(), err)
	assert.Equal(st.T(), int64(3), size, "Expected 3 as size")
}

func (st *ShifterTestSuit) TestShouldReturnRightNodePoolMaps() {

	kl := killer.NewKillerService(st.configMock, st.logger, st.k8sMock, st.gCloudMock, st.notifierMock)
	ss := NewShifterService(st.configMock, st.logger, st.k8sMock, st.gCloudMock, st.notifierMock, kl)

	nodePoolMap, err := ss.getNodePoolMap()
	assert.Nil(st.T(), err)

	expectedNPMap := map[string]npShiftConf{
		"services-np-1": {
			"services-p-1",
			1,
		},
	}

	assert.Equal(st.T(), true, assert.ObjectsAreEqual(nodePoolMap, expectedNPMap))

}

func (st *ShifterTestSuit) TestShouldShiftNodes() {

	ss := NewShifterService(st.configMock, st.logger, st.k8sMock, st.gCloudMock, st.notifierMock, st.killerMock)

	// st.k8sMock.AssertExpectations()
	st.gCloudMock.On("GetNumberOfZones").Return(3)
	onDemandNodes := v1.NodeList{
		Items: []v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-np-1-1",
					Labels: map[string]string{
						"component":                              "services",
						"criticality":                            "1",
						"failure-domain.beta.kubernetes.io/zone": "asia-south1-a",
						"cloud.google.com/gke-nodepool":          "services-np-1",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-np-1-2",
					Labels: map[string]string{
						"component":                              "services",
						"criticality":                            "1",
						"failure-domain.beta.kubernetes.io/zone": "asia-south1-b",
						"cloud.google.com/gke-nodepool":          "services-np-1",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-np-1-3",
					Labels: map[string]string{
						"component":                              "services",
						"criticality":                            "1",
						"failure-domain.beta.kubernetes.io/zone": "asia-south1-c",
						"cloud.google.com/gke-nodepool":          "services-np-1",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-np-1-4",
					Labels: map[string]string{
						"component":                              "services",
						"criticality":                            "1",
						"failure-domain.beta.kubernetes.io/zone": "asia-south1-a",
						"cloud.google.com/gke-nodepool":          "services-np-1",
					},
				},
			},
		},
	}
	preemptibleNodeList := v1.NodeList{
		Items: []v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "services-p-1-1",
					Labels: map[string]string{
						"component":                              "services",
						"criticality":                            "1",
						"failure-domain.beta.kubernetes.io/zone": "asia-south1-a",
						"cloud.google.com/gke-nodepool":          "services-p-1",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "services-p-1-2",
					Labels: map[string]string{
						"component":                              "services",
						"criticality":                            "1",
						"failure-domain.beta.kubernetes.io/zone": "asia-south1-b",
						"cloud.google.com/gke-nodepool":          "services-p-1",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "services-p-1-3",
					Labels: map[string]string{
						"component":                              "services",
						"criticality":                            "1",
						"failure-domain.beta.kubernetes.io/zone": "asia-south1-c",
						"cloud.google.com/gke-nodepool":          "services-p-1",
					},
				},
			},
		},
	}
	st.k8sMock.On("GetNodes", "cloud.google.com/gke-nodepool=services-np-1").Return(&onDemandNodes, nil)
	st.k8sMock.On("GetNodes", "cloud.google.com/gke-nodepool=services-p-1").Return(&preemptibleNodeList, nil)
	st.k8sMock.On("DeleteNode", mock.Anything).Return(nil)
	st.configMock.On("GetDuration", config.ShifterNPResizeTimeout).Return(10 * time.Minute)
	st.configMock.On("GetDuration", config.KillerDrainingTimeoutWhenNodeExpired).Return(time.Second)
	st.configMock.On("GetDuration", config.KillerEvictDeleteDeadline).Return(time.Millisecond)
	st.configMock.On("GetInt", config.KillerGracePeriodSecondsWhenPodDeleted).Return(1)
	st.configMock.On("GetDuration", config.ShifterSleepAfterNodeDeletion).Return(time.Second)
	st.gCloudMock.On("SetNodePoolSize", "services-p-1", int64(2), 10*time.Minute).Return(nil)

	for _, node := range onDemandNodes.Items {

		st.k8sMock.On("GetNode", node.Name).Return(node, nil)
		node.Spec.Unschedulable = true
		st.k8sMock.On("UpdateNode", node).Return(nil)
	}
	st.killerMock.On("EvictPodsFromNode", mock.AnythingOfType("string"), time.Second, time.Millisecond, 1).Return(nil)

	ss.shift()

	st.gCloudMock.AssertNumberOfCalls(st.T(), "ListNodePools", 1)
	st.k8sMock.AssertNumberOfCalls(st.T(), "GetNodes", 4)
	st.k8sMock.AssertNumberOfCalls(st.T(), "GetNode", 4)
	st.k8sMock.AssertNumberOfCalls(st.T(), "UpdateNode", 4)

	st.k8sMock.AssertExpectations(st.T())
	st.gCloudMock.AssertExpectations(st.T())
	st.killerMock.AssertExpectations(st.T())

}

func TestShiftererTestSuite(t *testing.T) {
	suite.Run(t, new(ShifterTestSuit))
}
