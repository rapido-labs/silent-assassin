package killer

import (
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *KillerTestSuite) TestShouldMakeNodeUnschedulable() {

	node := v1.Node{
		Spec: v1.NodeSpec{
			Unschedulable: false,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "Node-1",
			Annotations: map[string]string{"silent-assassin/expiry-time": time.Now().Add(time.Minute * -2).Format(time.RFC1123Z)}}}

	expectedNode := node.DeepCopy()
	expectedNode.Spec.Unschedulable = true

	k.k8sMock.On("UpdateNode", *expectedNode).Return(nil)

	ks := NewKillerService(k.configMock, k.logger, k.k8sMock, k.gCloudMock, k.notifierMock)

	err := ks.makeNodeUnschedulable(node)
	assert.Nil(k.T(), err)
	k.k8sMock.AssertExpectations(k.T())
}

func (k *KillerTestSuite) TestShouldStartNodeDrain() {

	pods := []v1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod1",
				Namespace: "ns1",
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: "Deployment",
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod2",
				Namespace: "ns2",
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: "Deployment",
					},
				},
			},
		},
	}

	k.k8sMock.On("GetPodsInNode", "Node-1").Return(pods, nil)
	k.k8sMock.On("DeletePod", "pod1", "ns1").Return(nil)
	k.k8sMock.On("DeletePod", "pod2", "ns2").Return(nil)

	ks := NewKillerService(k.configMock, k.logger, k.k8sMock, k.gCloudMock, k.notifierMock)

	err := ks.startNodeDrain("Node-1")
	assert.Nil(k.T(), err)
	k.k8sMock.AssertExpectations(k.T())
}

func (k *KillerTestSuite) TestShouldNotDrainDaemonsetPods() {

	pods := []v1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod1",
				Namespace: "ns1",
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: "DaemonSet",
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod2",
				Namespace: "ns2",
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: "Deployment",
					},
				},
			},
		},
	}

	k.k8sMock.On("GetPodsInNode", "Node-1").Return(pods, nil)
	k.k8sMock.On("DeletePod", "pod2", "ns2").Return(nil)

	ks := NewKillerService(k.configMock, k.logger, k.k8sMock, k.gCloudMock, k.notifierMock)

	err := ks.startNodeDrain("Node-1")
	assert.Nil(k.T(), err)
	k.k8sMock.AssertExpectations(k.T())
}

func (k *KillerTestSuite) TestShouldWaitforDrainingOfnodesWithTimeout() {
	nodeName := "node-1"
	pod1 := v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod-1", OwnerReferences: []metav1.OwnerReference{{Kind: "ReplicaSet"}}}}
	pod2 := v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod-2"}}
	pod3 := v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod-3"}}
	pod4 := v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod-4"}}

	k.k8sMock.On("GetPodsInNode", nodeName).Return([]v1.Pod{pod1, pod2, pod3, pod4}, nil).After(time.Second).Once()
	k.k8sMock.On("GetPodsInNode", nodeName).Return([]v1.Pod{pod1, pod2, pod3}, nil).After(time.Second).Once()
	k.k8sMock.On("GetPodsInNode", nodeName).Return([]v1.Pod{pod1, pod2}, nil).After(time.Second).Once()
	k.k8sMock.On("GetPodsInNode", nodeName).Return([]v1.Pod{pod1}, nil).After(time.Second).Once()
	k.k8sMock.On("GetPodsInNode", nodeName).Return([]v1.Pod{}, nil).After(time.Second).Once()

	ks := NewKillerService(k.configMock, k.logger, k.k8sMock, k.gCloudMock, k.notifierMock)
	assert.Nil(k.T(), ks.waitforDrainToFinish(nodeName, 5*time.Second), "err should be nothing")

	k.k8sMock.On("GetPodsInNode", nodeName).Return([]v1.Pod{pod1}, nil).After(2 * time.Second).Once()

	assert.NotNil(k.T(), ks.waitforDrainToFinish(nodeName, time.Second), "error should be something")
	k.k8sMock.AssertExpectations(k.T())
}

func (k *KillerTestSuite) TestShouldTriggerEvacuationOfPodsFromNode() {

	node := v1.Node{
		Spec: v1.NodeSpec{
			Unschedulable: false,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "Node-1",
			Annotations: map[string]string{"silent-assassin/expiry-time": time.Now().Add(time.Minute * -2).Format(time.RFC1123Z)}}}

	pods := []v1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod1",
				Namespace: "ns1",
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: "Deployment",
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod2",
				Namespace: "ns2",
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: "Deployment",
					},
				},
			},
		},
	}

	expectedNode := node.DeepCopy()
	expectedNode.Spec.Unschedulable = true

	k.k8sMock.On("GetNode", "Node-1").Return(node, nil)
	k.k8sMock.On("UpdateNode", *expectedNode).Return(nil)
	k.k8sMock.On("GetPodsInNode", "Node-1").Return(pods, nil).Once()
	k.k8sMock.On("GetPodsInNode", "Node-1").Return([]v1.Pod{}, nil).Once()
	k.k8sMock.On("DeletePod", "pod1", "ns1").Return(nil)
	k.k8sMock.On("DeletePod", "pod2", "ns2").Return(nil)

	ks := NewKillerService(k.configMock, k.logger, k.k8sMock, k.gCloudMock, k.notifierMock)

	assert.Nil(k.T(), ks.EvacuatePodsFromNode("Node-1", 10, true), "Error was not expected")
}
