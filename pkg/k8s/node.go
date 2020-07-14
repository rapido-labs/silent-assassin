package k8s

import (
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kc KubernetesClient) GetNodes(labels []string) *v1.NodeList {

	labelSelector := strings.Join(labels, ",")

	kc.logger.Debug(fmt.Sprintf("Label Selectors : %v", labelSelector))

	options := metav1.ListOptions{
		LabelSelector: labelSelector,
	}
	nodes, err := kc.CoreV1().Nodes().List(options)

	if err != nil {
		kc.logger.Info("Failed to get nodes")
		panic(err)
	}

	return nodes
}

func (kc KubernetesClient) GetNode(name string) (v1.Node, error) {
	options := metav1.GetOptions{}

	node, err := kc.CoreV1().Nodes().Get(name, options)
	return *node, err
}

func (kc KubernetesClient) AnnotateNode(node v1.Node) error {

	_, err := kc.CoreV1().Nodes().Update(&node)
	return err
}

func (kc KubernetesClient) UpdateNode(node v1.Node) error {
	_, err := kc.CoreV1().Nodes().Update(&node)
	return err
}

func (kc KubernetesClient) DeleteNode(name string) error {
	options := &metav1.DeleteOptions{}
	err := kc.CoreV1().Nodes().Delete(name, options)
	return err
}
