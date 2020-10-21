package gcloud

import (
	"context"

	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
)

type GCloudClient struct {
	Service *compute.Service
}

type IGCloudClient interface {
	DeleteInstance(projectID, zone, name string) error
}

func NewClient() IGCloudClient {
	var gClient IGCloudClient

	client, err := google.DefaultClient(context.Background(), compute.ComputeScope)
	if err != nil {
		panic(err.Error())
	}

	service, err := compute.New(client)

	if err != nil {
		return gClient
	}

	gClient = GCloudClient{
		Service: service,
	}

	return gClient
}

func (client GCloudClient) DeleteInstance(projectID, zone, name string) error {
	_, err := client.Service.Instances.Delete(projectID, zone, name).Context(context.Background()).Do()
	return err
}
