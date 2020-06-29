package killer

import (
	"testing"

	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/gcloud"
	"github.com/roppenlabs/silent-assassin/pkg/k8s"
	"github.com/roppenlabs/silent-assassin/pkg/logger"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
)

func TestShouldFetchNodesWithLabels(t *testing.T) {
	k8sMock := new(k8s.K8sClientMock)
	configMock := new(config.ProviderMock)
	gCloudClientMock := new(gcloud.GCloudClientMock)

	configMock.On("GetString", mock.Anything).Return("debug")
	configMock.On("GetInt", "killer.killer.poll_interval_ms")
	configMock.On("GetStringSlice", "spotter.label_selectors").Return([]string{"cloud.google.com/gke-preemptible=true,label2=test"})
	k8sMock.On("GetNodes", []string{"cloud.google.com/gke-preemptible=true,label2=test"}).Return(&v1.NodeList{})

	zapLogger := logger.Init(configMock)
	kubeClient := k8sMock
	gcloudClient := gCloudClientMock
	ss := NewKillerService(configMock, zapLogger, kubeClient, gcloudClient)

	ss.kill()

	k8sMock.AssertExpectations(t)
}
