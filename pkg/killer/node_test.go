package killer

import (
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/roppenlabs/silent-assassin/pkg/k8s"
)

func (k *KillerTestSuite) setupNodeToDrain(name string) {
	k.T().Helper()

	node := corev1.Node{
		Spec: corev1.NodeSpec{
			Unschedulable: false,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: map[string]string{"silent-assassin/expiry-time": time.Now().Add(time.Minute * -2).Format(time.RFC1123Z)},
		},
	}

	k.k8sMock.On("GetNode", name).Return(node, nil)
	expectedNode := node.DeepCopy()
	expectedNode.Spec.Unschedulable = true
	k.k8sMock.On("UpdateNode", *expectedNode).Return(nil)
}

func (k *KillerTestSuite) TestShouldMakeNodeUnschedulable() {

	node := corev1.Node{
		Spec: corev1.NodeSpec{
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

// TestDeletePodsFromNode tests delete pods from node
func (k *KillerTestSuite) TestDeletePodsFromNode() {
	nodeName := "node-1"
	timeout := 2 * time.Second
	gracePeriodSeconds := 1

	k.setupNodeToDrain(nodeName)

	k.k8sMock.On("DrainNode", nodeName, false, timeout, gracePeriodSeconds).Return(nil).Once()
	ks := NewKillerService(k.configMock, k.logger, k.k8sMock, k.gCloudMock, k.notifierMock)
	k.Assert().NoError(ks.DeletePodsFromNode(nodeName, timeout, gracePeriodSeconds))
	k.k8sMock.AssertExpectations(k.T())
}

// TestEvictPodsFromNodeNoTimeout tests evict pods from node, evict not timeout, delete is not called
func (k *KillerTestSuite) TestEvictPodsFromNodeNoTimeout() {
	nodeName := "node-1"
	timeout := 10 * time.Second
	evictDeleteDeadline := 2 * time.Second
	gracePeriodSeconds := 1
	evictTimeout := timeout - evictDeleteDeadline
	k.setupNodeToDrain(nodeName)

	k.k8sMock.On("DrainNode", nodeName, true, evictTimeout, gracePeriodSeconds).Return(nil).Once()

	ks := NewKillerService(k.configMock, k.logger, k.k8sMock, k.gCloudMock, k.notifierMock)
	k.Assert().NoError(ks.EvictPodsFromNode(nodeName, timeout, evictDeleteDeadline, gracePeriodSeconds))
	k.k8sMock.AssertExpectations(k.T())
}

// TestEvictPodsFromNodeTimeout tests evict pods from node, evict timeout, check delete is called
func (k *KillerTestSuite) TestEvictPodsFromNodeTimeout() {
	nodeName := "node-1"
	timeout := 10 * time.Second
	evictDeleteDeadline := 2 * time.Second
	gracePeriodSeconds := 1
	evictTimeout := timeout - evictDeleteDeadline
	k.setupNodeToDrain(nodeName)

	k.k8sMock.On("DrainNode", nodeName, true, evictTimeout, gracePeriodSeconds).Return(k8s.ErrDrainNodeTimeout).Once()
	k.k8sMock.On("DrainNode", nodeName, false, evictDeleteDeadline, gracePeriodSeconds).Return(nil).Once()

	ks := NewKillerService(k.configMock, k.logger, k.k8sMock, k.gCloudMock, k.notifierMock)
	k.Assert().NoError(ks.EvictPodsFromNode(nodeName, timeout, evictDeleteDeadline, gracePeriodSeconds))
	k.k8sMock.AssertExpectations(k.T())
}
