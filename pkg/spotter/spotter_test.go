package spotter

import (
	"testing"
	"time"

	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/k8s"
	"github.com/roppenlabs/silent-assassin/pkg/logger"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestShouldFetchNodesWithLabels(t *testing.T) {

	k8sMock := new(k8s.K8sClientMock)
	configMock := new(config.ProviderMock)

	configMock.On("GetString", mock.Anything).Return("debug")
	configMock.On("GetInt", "spotter.poll_interval_ms").Return(10)
	configMock.On("GetStringSlice", "spotter.label_selectors").Return([]string{"cloud.google.com/gke-preemptible=true,label2=test"})
	k8sMock.On("GetNodes", []string{"cloud.google.com/gke-preemptible=true,label2=test"}).Return(&v1.NodeList{})

	zapLogger := logger.Init(configMock)
	kubeClient := k8sMock
	ss := NewSpotterService(configMock, zapLogger, kubeClient)

	ss.spot()

	k8sMock.AssertExpectations(t)
}

func TestShouldAnnotateIfAbsent(t *testing.T) {

	k8sMock := new(k8s.K8sClientMock)
	configMock := new(config.ProviderMock)

	nodeAlreadyAnnotated := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "Node-1",
			Annotations: map[string]string{"silent-assassin/expiry-time": time.Now().String()}}}

	nodeToBeAnnotated := v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "Node-2"}}

	nodeList := v1.NodeList{
		Items: []v1.Node{nodeAlreadyAnnotated, nodeToBeAnnotated},
	}

	configMock.On("GetString", mock.Anything).Return("debug")
	configMock.On("GetInt", mock.Anything).Return(10)
	configMock.On("GetStringSlice", mock.Anything).Return([]string{"cloud.google.com/gke-preemptible=true,label2=test"})
	k8sMock.On("GetNodes", mock.Anything).Return(&nodeList)
	k8sMock.On("AnnotateNode", mock.MatchedBy(func(input v1.Node) bool {
		_, found := input.ObjectMeta.Annotations["silent-assassin/expiry-time"]
		return found && input.ObjectMeta.Name == "Node-2"
	})).Return(nil)

	zapLogger := logger.Init(configMock)
	kubeClient := k8sMock
	ss := NewSpotterService(configMock, zapLogger, kubeClient)

	ss.spot()

	k8sMock.AssertExpectations(t)
}

func TestShouldSetExpiryTimeAs12HoursFromCreation(t *testing.T) {

	k8sMock := new(k8s.K8sClientMock)
	configMock := new(config.ProviderMock)

	creationTimestamp, _ := time.Parse(time.RFC1123Z, "Mon, 22 Jun 2020 12:54:45 +0530")

	nodeAlreadyAnnotated := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "Node-1",
			CreationTimestamp: metav1.NewTime(creationTimestamp),
			Annotations:       map[string]string{"silent-assassin/expiry-time": time.Now().String()}}}

	nodeToBeAnnotated := v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "Node-2", CreationTimestamp: metav1.NewTime(creationTimestamp)}}

	nodeList := v1.NodeList{
		Items: []v1.Node{nodeAlreadyAnnotated, nodeToBeAnnotated},
	}

	configMock.On("GetString", mock.Anything).Return("debug")
	configMock.On("GetInt", mock.Anything).Return(10)
	configMock.On("GetStringSlice", mock.Anything).Return([]string{"cloud.google.com/gke-preemptible=true,label2=test"})
	k8sMock.On("GetNodes", mock.Anything).Return(&nodeList)
	k8sMock.On("AnnotateNode", mock.MatchedBy(func(input v1.Node) bool {
		expiryTimeAsString, found := input.ObjectMeta.Annotations["silent-assassin/expiry-time"]
		expiryTime, _ := time.Parse(time.RFC1123Z, expiryTimeAsString)
		return found && input.ObjectMeta.Name == "Node-2" && expiryTime == creationTimestamp.Add(time.Hour*12)
	})).Return(nil)

	zapLogger := logger.Init(configMock)
	kubeClient := k8sMock
	ss := NewSpotterService(configMock, zapLogger, kubeClient)

	ss.spot()

	k8sMock.AssertExpectations(t)
}
