package killer

import (
	"testing"
	"time"

	"github.com/alecthomas/assert"
	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/gcloud"
	"github.com/roppenlabs/silent-assassin/pkg/k8s"
	"github.com/roppenlabs/silent-assassin/pkg/logger"
	"github.com/roppenlabs/silent-assassin/pkg/notifier"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KillerTypeSuite struct {
	suite.Suite
	k8sMock      *k8s.K8sClientMock
	gCloudMock   *gcloud.GCloudClientMock
	configMock   *config.ProviderMock
	logger       logger.IZapLogger
	notifierMock *notifier.NotifierClientMock
}

func (k *KillerTypeSuite) SetupTest() {
	k.configMock = new(config.ProviderMock)
	k.k8sMock = new(k8s.K8sClientMock)
	k.gCloudMock = new(gcloud.GCloudClientMock)
	k.notifierMock = new(notifier.NotifierClientMock)
	k.notifierMock.On("Info", mock.Anything, mock.Anything)
	k.notifierMock.On("Error", mock.Anything, mock.Anything)
	k.configMock.On("GetString", mock.Anything).Return("debug")
	k.configMock.On("GetStringSlice", "spotter.label_selectors").Return([]string{"cloud.google.com/gke-preemptible=true,label2=test"})
	k.logger = logger.Init(k.configMock)
}
func (k *KillerTypeSuite) TestShouldReturnExpiredNodes() {
	preemptibleNodeExpired := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "Node-1",
			Annotations: map[string]string{"silent-assassin/expiry-time": time.Now().Add(time.Minute * -2).Format(time.RFC1123Z)}}}
	preemptibleNodeNotExpired := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "Node-2",
			Annotations: map[string]string{"silent-assassin/expiry-time": time.Now().Add(time.Hour * 2).Format(time.RFC1123Z)}}}
	nodeList := v1.NodeList{
		Items: []v1.Node{preemptibleNodeExpired, preemptibleNodeNotExpired}}

	k.k8sMock.On("GetNodes", []string{"cloud.google.com/gke-preemptible=true,label2=test"}).Return(&nodeList)

	ks := NewKillerService(k.configMock, k.logger, k.k8sMock, k.gCloudMock, k.notifierMock)
	ks.findExpiredTimeNodes([]string{"cloud.google.com/gke-preemptible=true,label2=test"})

	assert.Contains(k.T(), ks.findExpiredTimeNodes([]string{"cloud.google.com/gke-preemptible=true,label2=test"}), preemptibleNodeExpired, "Node-1 should be returned")
	assert.NotContains(k.T(), ks.findExpiredTimeNodes([]string{"cloud.google.com/gke-preemptible=true,label2=test"}), preemptibleNodeNotExpired, "Node-2 should not be returned")
	k.k8sMock.AssertExpectations(k.T())
}

func (k *KillerTypeSuite) TestShouldFilterPodsByReferenceKind() {
	podOwnedByDS := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			OwnerReferences: []metav1.OwnerReference{
				{Kind: "DaemonSet"}}}}
	podOwnedByRS := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			OwnerReferences: []metav1.OwnerReference{
				{Kind: "ReplicaSet"}}}}
	podList := []v1.Pod{podOwnedByDS, podOwnedByRS}
	filteredPodList := filterOutPodByOwnerReferenceKind(podList, "DaemonSet")

	assert.Contains(k.T(), filteredPodList, podOwnedByRS)
	assert.NotContains(k.T(), filteredPodList, podOwnedByDS)
}

func (k *KillerTypeSuite) TestShouldWaitforDrainingOfnodesWithTimeout() {
	nodeName := "node-1"
	pod1 := v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod-1", OwnerReferences: []metav1.OwnerReference{{Kind: "ReplicaSet"}}}}
	pod2 := v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod-2"}}
	pod3 := v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod-3"}}
	pod4 := v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod-4"}}

	k.k8sMock.On("GetPodsInNode", nodeName).Return([]v1.Pod{pod1, pod2, pod3, pod4}, nil).After(1000 * time.Millisecond).Once()
	k.k8sMock.On("GetPodsInNode", nodeName).Return([]v1.Pod{pod1, pod2, pod3}, nil).After(1000 * time.Millisecond).Once()
	k.k8sMock.On("GetPodsInNode", nodeName).Return([]v1.Pod{pod1, pod2}, nil).After(1000 * time.Millisecond).Once()
	k.k8sMock.On("GetPodsInNode", nodeName).Return([]v1.Pod{pod1}, nil).After(1000 * time.Millisecond).Once()
	k.k8sMock.On("GetPodsInNode", nodeName).Return([]v1.Pod{}, nil).After(1000 * time.Millisecond).Once()

	ks := NewKillerService(k.configMock, k.logger, k.k8sMock, k.gCloudMock, k.notifierMock)
	assert.Nil(k.T(), ks.waitforDrainToFinish(nodeName, 5000), "err should be nothing")

	k.k8sMock.On("GetPodsInNode", nodeName).Return([]v1.Pod{pod1}, nil).After(2000 * time.Millisecond).Once()

	assert.NotNil(k.T(), ks.waitforDrainToFinish(nodeName, 1000), "error should be something")
	k.k8sMock.AssertExpectations(k.T())
}

func TestKillerTestSuite(t *testing.T) {
	suite.Run(t, new(KillerTypeSuite))
}
