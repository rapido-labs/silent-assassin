package k8s

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubectl/pkg/drain"

	"github.com/roppenlabs/silent-assassin/pkg/logger"
)

// GetNodes gets a list of nodes using given label selector.
func (kc KubernetesClient) GetNodes(labelSelector string) (*corev1.NodeList, error) {

	kc.logger.Debug(fmt.Sprintf("Label Selectors : %v", labelSelector))

	options := metav1.ListOptions{
		LabelSelector: labelSelector,
	}
	nodes, err := kc.CoreV1().Nodes().List(kc.ctx, options)

	if err != nil {
		kc.logger.Info("Failed to get nodes")
		return nodes, err
	}

	return nodes, err
}

// GetNode gets node object using a node name.
func (kc KubernetesClient) GetNode(name string) (corev1.Node, error) {
	options := metav1.GetOptions{}

	node, err := kc.CoreV1().Nodes().Get(kc.ctx, name, options)
	return *node, err
}

// UpdateNode updates a node object.
func (kc KubernetesClient) UpdateNode(node corev1.Node) error {
	_, err := kc.CoreV1().Nodes().Update(kc.ctx, &node, metav1.UpdateOptions{})
	return err
}

// DeleteNode deletes a node from cluster.
func (kc KubernetesClient) DeleteNode(name string) error {
	options := metav1.DeleteOptions{}
	err := kc.CoreV1().Nodes().Delete(kc.ctx, name, options)
	return err
}

// DrainNode drains pods from a node. It deletes or evicts pods from the node , depending on useEvict.
// DrainNode executes synchronously, it blocks until the node is either fully drained, or timeout is reached.
// gracePeriodSeconds is used as drain.Helper.GracePeriodSeconds, if negative, the default
// value specified in the pod will be used.
func (kc KubernetesClient) DrainNode(name string, useEvict bool, timeout time.Duration, gracePeriodSeconds int) error {
	drainer := &drain.Helper{
		Ctx:                 kc.ctx,
		Client:              kc,
		GracePeriodSeconds:  gracePeriodSeconds,
		IgnoreAllDaemonSets: true,
		DeleteEmptyDirData:  true,
		Out:                 logger.LogWriter{LogFunc: kc.logger.Info},
		ErrOut:              logger.LogWriter{LogFunc: kc.logger.Error},
		OnPodDeletedOrEvicted: func(pod *corev1.Pod, usingEviction bool) {
			verb := "deleted"
			if useEvict {
				verb = "evicted"
			}
			kc.logger.Info(fmt.Sprintf("pod %s %s", pod.Name, verb))
		},

		DisableEviction: !useEvict,
		Timeout:         timeout,
	}

	podList, errs := drainer.GetPodsForDeletion(name)
	if errs != nil {
		return fmt.Errorf("%s", errs)
	}

	err := drainer.DeleteOrEvictPods(podList.Pods())
	if err != nil {
		// XXX: the drain package returned error is not explicitly typed,
		// so unfortunally we have to use text match to check the timeout condition.
		// Timeout can happen in 2 cases:
		// 1. some misconfigured PDB causing pods never able to be evicted
		// 2. pod graceful termination is too long that it breaks drainer timeout
		errorText := err.Error()
		timeoutWaitingForPodEviction := strings.Contains(errorText, "timed out waiting for the condition")
		timeoutWaitingForPodGracefulTermination := strings.Contains(errorText, "global timeout reached")
		if timeoutWaitingForPodGracefulTermination || timeoutWaitingForPodEviction {
			return errors.Wrap(ErrDrainNodeTimeout, errorText)
		}
	}
	return err
}
