package k8s

import (
	"fmt"
	"strings"

	"github.com/roppenlabs/silent-assassin/pkg/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kc KubernetesClient) GetNodes(labels []string) *v1.NodeList {

	labelSelector := strings.Join(labels, ",")

	kc.logger.Info(fmt.Sprintf("Label Selectors : %v", labelSelector), "GET_NODES", config.LogComponentName)

	options := metav1.ListOptions{
		LabelSelector: labelSelector,
	}
	nodes, err := kc.CoreV1().Nodes().List(options)

	if err != nil {
		kc.logger.Info("Failed to get nodes", "GET_NODES", config.LogComponentName)
		panic(err)
	}

	return nodes
}

func (kc KubernetesClient) AnnotateNode(node v1.Node) error {

	_, err := kc.CoreV1().Nodes().Update(&node)
	return err
}
