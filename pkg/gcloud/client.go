package gcloud

import (
	"context"
	"fmt"
	"time"

	"github.com/roppenlabs/silent-assassin/pkg/k8s"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
	container "google.golang.org/api/container/v1"
)

type GCloudClient struct {
	project          string
	location         string
	cluster          string
	kubeClient       k8s.IKubernetesClient
	computeService   *compute.Service
	containerService *container.Service
}

type IGCloudClient interface {
	DeleteInstance(zone, name string) error
	GetInstance(project, zone, name string) (*compute.Instance, error)
	ListNodePools() ([]*container.NodePool, error)
	GetNodePool(npName string) (*container.NodePool, error)
	SetNodePoolSize(npName string, size int64, timeout int) error
}

func NewClient(kc k8s.IKubernetesClient) IGCloudClient {
	var gClient GCloudClient

	computeClient, err := google.DefaultClient(context.Background(), compute.ComputeScope)
	if err != nil {
		panic(err.Error())
	}

	computeService, err := compute.New(computeClient)

	if err != nil {
		panic(err.Error())
	}

	containerClient, err := google.DefaultClient(context.Background(), container.CloudPlatformScope)
	if err != nil {
		panic(err.Error())
	}

	containerService, err := container.New(containerClient)

	if err != nil {
		panic(err.Error())
	}

	gClient = GCloudClient{
		computeService:   computeService,
		containerService: containerService,
		kubeClient:       kc,
	}

	// gClient.project, gClient.location, gClient.cluster, _ = gClient.setclusterDetails()
	err = gClient.setClusterDetails()

	if err != nil {
		panic(err.Error())
	}

	return gClient
}

func (client GCloudClient) DeleteInstance(zone, name string) error {
	_, err := client.computeService.Instances.Delete(client.project, zone, name).Context(context.Background()).Do()
	return err
}

func (client GCloudClient) GetInstance(project, zone, name string) (*compute.Instance, error) {
	instance, err := client.computeService.Instances.Get(project, zone, name).Context(context.Background()).Do()
	if err != nil {
		return instance, err
	}
	return instance, err
}

func (client GCloudClient) ListNodePools() ([]*container.NodePool, error) {

	clusterURI := fmt.Sprintf("projects/%s/locations/%s/clusters/%s", client.project, client.location, client.cluster)
	npsResp, err := client.containerService.Projects.Locations.Clusters.NodePools.List(clusterURI).Context(context.Background()).Do()

	if err != nil {
		return nil, err
	}

	return npsResp.NodePools, err
}

func (client GCloudClient) GetNodePool(name string) (*container.NodePool, error) {

	npURI := fmt.Sprintf("projects/%s/locations/%s/clusters/%s/nodePools/%s", client.project, client.location, client.cluster, name)
	np, err := client.containerService.Projects.Locations.Clusters.NodePools.Get(npURI).Context(context.Background()).Do()

	if err != nil {
		return nil, err
	}
	return np, nil
}

func (client GCloudClient) waitForLocationOperation(ctx context.Context, operationID string) error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	opURI := fmt.Sprintf("projects/%s/locations/%s/operations/%s", client.project, client.location, operationID)
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for operation to complete")
		case <-ticker.C:
			result, err := client.containerService.Projects.Locations.Operations.Get(opURI).Context(context.Background()).Do()
			if err != nil {
				return fmt.Errorf("Locations.Operations.Get: %s", err)
			}

			if result.Status == "DONE" {
				return nil
			}
		}
	}
}

func (client GCloudClient) SetNodePoolSize(name string, size int64, timeout int) error {

	npURI := fmt.Sprintf("projects/%s/locations/%s/clusters/%s/nodePools/%s", client.project, client.location, client.cluster, name)

	npResizeRequest := &container.SetNodePoolSizeRequest{
		NodeCount: size,
	}

	op, err := client.containerService.Projects.Locations.Clusters.NodePools.SetSize(npURI, npResizeRequest).Context(context.Background()).Do()

	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Minute)
	defer cancel()

	err = client.waitForLocationOperation(ctx, op.Name)

	if err != nil {
		return err
	}
	return nil
}
