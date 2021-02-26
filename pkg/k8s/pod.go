package k8s

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kc KubernetesClient) GetPodsInNode(name string) ([]v1.Pod, error) {
	pods := []v1.Pod{}
	options := metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", name),
	}
	podList, err := kc.CoreV1().Pods("").List(kc.ctx, options)
	if err != nil {
		return pods, err
	}
	pods = podList.Items

	return pods, err

}
