package gcloud

import "cloud.google.com/go/compute/metadata"

type IMetadata interface {
	InstanceName() (string, error)
	Subscribe(suffix string, fn func(v string, ok bool) error) error
}
type Mclient struct {
}

func (m Mclient) InstanceName() (string, error) {
	return metadata.InstanceName()
}

func (m Mclient) Subscribe(suffix string, fn func(v string, ok bool) error) error {
	return metadata.Subscribe(suffix, fn)
}
