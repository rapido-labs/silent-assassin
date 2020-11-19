package gcloud

import (
	"errors"
	"strings"
)

func (client *GCloudClient) setClusterDetails() error {
	var project, location, cluster string

	nodes, err := client.kubeClient.GetNodes("")

	if err != nil {
		return err
	}

	if len(nodes.Items) == 0 {
		return errors.New("zero nodes in the cluster")
	}

	s := strings.Split(nodes.Items[0].Spec.ProviderID, "/")
	project = s[2]
	zone := s[3]
	name := s[4]

	instance, err := client.GetInstance(project, zone, name)

	if err != nil {
		return err
	}

	for _, metadata := range instance.Metadata.Items {
		if metadata.Key == "cluster-name" {
			cluster = *metadata.Value
		}
		if metadata.Key == "cluster-location" {
			location = *metadata.Value
		}
		if cluster != "" && location != "" {
			break
		}
	}

	client.project = project
	client.location = location
	client.cluster = cluster

	return nil
}
