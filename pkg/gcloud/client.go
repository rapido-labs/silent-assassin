package gcloud

import (
	"context"

	"github.com/roppenlabs/silent-assassin/pkg/logger"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
)

type GCloudClient struct {
	Service *compute.Service
	logger  logger.IZapLogger
}

type IGCloudClient interface {
	DeleteNode(projectID, zone, name string) error
}

func NewClient(zl logger.IZapLogger) (IGCloudClient, error) {
	var gClient IGCloudClient

	client, err := google.DefaultClient(context.Background(), compute.ComputeScope)
	if err != nil {
		return gClient, err
	}

	service, err := compute.New(client)

	if err != nil {
		return gClient, err
	}

	gClient = GCloudClient{
		Service: service,
	}

	return gClient, err
}

func (client GCloudClient) DeleteNode(projectID, zone, name string) error {
	_, err := client.Service.Instances.Delete(projectID, zone, name).Context(context.Background()).Do()
	return err
}
