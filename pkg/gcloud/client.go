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
	project                    string
	location                   string
	cluster                    string
	kubeClient                 k8s.IKubernetesClient
	computeServiceComputeScope *compute.Service
	containerServiceCloudScope *container.Service
	computeServiceCloudScope   *compute.Service
}

type IGCloudClient interface {
	DeleteInstance(zone, name string) error
	GetInstance(project, zone, name string) (*compute.Instance, error)
	ListNodePools() ([]*container.NodePool, error)
	GetNodePool(npName string) (*container.NodePool, error)
	SetNodePoolSize(npName string, size int64, timeout time.Duration) error
	GetNumberOfZones() int
}

func NewClient(kc k8s.IKubernetesClient) IGCloudClient {
	var gClient GCloudClient

	computeClientComputeScope, err := google.DefaultClient(context.Background(), compute.ComputeScope)
	if err != nil {
		panic(err.Error())
	}

	computeServiceComputeScope, err := compute.New(computeClientComputeScope)

	if err != nil {
		panic(err.Error())
	}

	containerClientCloudScope, err := google.DefaultClient(context.Background(), container.CloudPlatformScope)
	if err != nil {
		panic(err.Error())
	}

	containerServiceCloudScope, err := container.New(containerClientCloudScope)

	if err != nil {
		panic(err.Error())
	}

	computeClientCloudScope, err := google.DefaultClient(context.Background(), compute.CloudPlatformScope)

	if err != nil {
		panic(err.Error())
	}

	computeServiceCloudScope, err := compute.New(computeClientCloudScope)

	if err != nil {
		panic(err.Error())
	}
	gClient = GCloudClient{
		computeServiceComputeScope: computeServiceComputeScope,
		containerServiceCloudScope: containerServiceCloudScope,
		computeServiceCloudScope:   computeServiceCloudScope,
		kubeClient:                 kc,
	}

	// gClient.project, gClient.location, gClient.cluster, _ = gClient.setclusterDetails()
	err = gClient.setClusterDetails()

	if err != nil {
		panic(err.Error())
	}

	return gClient
}

func (client GCloudClient) DeleteInstance(zone, name string) error {
	_, err := client.computeServiceComputeScope.Instances.Delete(client.project, zone, name).Context(context.Background()).Do()
	return err
}

func (client GCloudClient) GetInstance(project, zone, name string) (*compute.Instance, error) {
	instance, err := client.computeServiceComputeScope.Instances.Get(project, zone, name).Context(context.Background()).Do()
	if err != nil {
		return instance, err
	}
	return instance, err
}

func (client GCloudClient) ListNodePools() ([]*container.NodePool, error) {

	clusterURI := fmt.Sprintf("projects/%s/locations/%s/clusters/%s", client.project, client.location, client.cluster)
	npsResp, err := client.containerServiceCloudScope.Projects.Locations.Clusters.NodePools.List(clusterURI).Context(context.Background()).Do()

	if err != nil {
		return nil, err
	}

	return npsResp.NodePools, err
}

func (client GCloudClient) GetNodePool(name string) (*container.NodePool, error) {

	npURI := fmt.Sprintf("projects/%s/locations/%s/clusters/%s/nodePools/%s", client.project, client.location, client.cluster, name)
	np, err := client.containerServiceCloudScope.Projects.Locations.Clusters.NodePools.Get(npURI).Context(context.Background()).Do()

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
			result, err := client.containerServiceCloudScope.Projects.Locations.Operations.Get(opURI).Context(context.Background()).Do()
			if err != nil {
				return fmt.Errorf("Locations.Operations.Get: %s", err)
			}

			if result.Status == "DONE" {
				return nil
			}
		}
	}
}

func (client GCloudClient) SetNodePoolSize(name string, size int64, timeout time.Duration) error {

	npURI := fmt.Sprintf("projects/%s/locations/%s/clusters/%s/nodePools/%s", client.project, client.location, client.cluster, name)

	npResizeRequest := &container.SetNodePoolSizeRequest{
		NodeCount: size,
	}

	op, err := client.containerServiceCloudScope.Projects.Locations.Clusters.NodePools.SetSize(npURI, npResizeRequest).Context(context.Background()).Do()

	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err = client.waitForLocationOperation(ctx, op.Name)

	if err != nil {
		return err
	}
	return nil
}

func (client GCloudClient) GetNumberOfZones() int {
	regionURI := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%v/regions/%v", client.project, client.location)
	filterByRegion := fmt.Sprintf("region = \"%v\"", regionURI)
	req := client.computeServiceCloudScope.Zones.List(client.project).Filter(filterByRegion)
	zoneCount := 0
	if err := req.Pages(context.Background(), func(page *compute.ZoneList) error {
		zoneCount += len(page.Items)
		return nil
	}); err != nil {
		panic(err)
	}
	return zoneCount
}
