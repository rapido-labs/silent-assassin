package killer

import (
	"errors"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/k8s"
)

// getExpiryTime returns the expiry time set on the node
// using the annotation silent-assassin/expiry-time
func getExpiryTime(node corev1.Node) string {

	if timestamp, ok := node.ObjectMeta.Annotations[config.ExpiryTimeAnnotation]; ok {
		return timestamp
	}

	return ""
}

// findExpiredTimeNodes gets the list of nodes whose expiry time set is older than current time
// These nodes are eligible for deletion.
func (ks KillerService) findExpiredTimeNodes(labelSelector string) ([]corev1.Node, error) {
	var nodesToBeDeleted []corev1.Node
	nodeList, err := ks.kubeClient.GetNodes(labelSelector)
	if err != nil {
		ks.logger.Error(fmt.Sprintf("Error getting nodes %s", err.Error()))
		ks.notifier.Error(config.EventGetNodes, fmt.Sprintf("Error fetching nodes %s", err.Error()))
		return nodesToBeDeleted, err
	}

	now := time.Now().UTC()

	for _, node := range nodeList.Items {
		timestamp := getExpiryTime(node)
		if timestamp == "" {
			ks.logger.Warn(fmt.Sprintf("Node %s does not have %s annotation set", node.Name, config.ExpiryTimeAnnotation))
			continue
		}
		expiryDatetime, err := time.Parse(time.RFC1123Z, timestamp)
		if err != nil {
			ks.logger.Error(fmt.Sprintf("Error parsing expiry datetime with value '%s', %s", expiryDatetime, err.Error()))
			continue
		}

		timeDiff := expiryDatetime.Sub(now).Minutes()

		if timeDiff <= 0 {
			nodesToBeDeleted = append(nodesToBeDeleted, node)
		}

	}

	return nodesToBeDeleted, nil
}

// makeNodeUnschedulable function cordons the node thus disabling scheduling of
// any new pods on this node during draining.
func (ks KillerService) makeNodeUnschedulable(node corev1.Node) error {

	node.Spec.Unschedulable = true
	err := ks.kubeClient.UpdateNode(node)
	return err

}

// getZoneFromNode extracts the GCP projectID and zone
// from the given node.
func getZoneFromNode(node corev1.Node) string {
	s := strings.Split(node.Spec.ProviderID, "/")
	zone := s[3]
	return zone
}

func (ks KillerService) deleteNode(node corev1.Node) {
	nodeDetails := getNodeDetails(node)

	// Delete the k8s node
	ks.logger.Info(fmt.Sprintf("Deleting node %s", node.Name))
	if err := ks.kubeClient.DeleteNode(node.Name); err != nil {
		ks.logger.Info(fmt.Sprintf("Error deleting the node %s", node.Name))
		ks.notifier.Error(config.EventDeleteNode, fmt.Sprintf("%s\nError:%s", nodeDetails, err.Error()))
		return
	}

	ks.notifier.Info(config.EventDeleteNode, nodeDetails)

	// Delete gcloud instance.
	zone := getZoneFromNode(node)
	ks.logger.Info(fmt.Sprintf("Deletig google instance %s", node.Name))
	if err := ks.gcloudClient.DeleteInstance(zone, node.Name); err != nil {
		ks.logger.Error(fmt.Sprintf("Could not kill the node %s %s", node.Name, err.Error()))
		ks.notifier.Error(config.EventDeleteInstance, fmt.Sprintf("%s\nError:%s", nodeDetails, err.Error()))
		return
	}
	ks.notifier.Info(config.EventDeleteInstance, nodeDetails)
}

func getNodeDetails(node corev1.Node) string {
	return fmt.Sprintf("Node: %s\n"+
		"Creation Time: %s\n"+
		"ExpiryTime: %s",
		node.Name, node.CreationTimestamp, node.Annotations[config.ExpiryTimeAnnotation])
}

// DeletePodsFromNode drains pods from a node using delete k8s delete api.
func (ks KillerService) DeletePodsFromNode(name string, timeout time.Duration, gracePeriodSeconds int) error {
	start := time.Now()

	node, err := ks.kubeClient.GetNode(name)

	if err != nil {
		ks.logger.Error(fmt.Sprintf("Error fetching the node %s, %s", name, err))
		return err
	}
	nodeDetails := getNodeDetails(node)

	if err := ks.makeNodeUnschedulable(node); err != nil {
		ks.logger.Error(fmt.Sprintf("Failed to cordon the node %s, %s", node.Name, err))
		ks.notifier.Error(config.EventCordon, fmt.Sprintf("%s\nError:%s", nodeDetails, err))
		return err
	}

	if err := ks.kubeClient.DrainNode(node.Name, false, timeout, gracePeriodSeconds); err != nil {
		ks.logger.Error(fmt.Sprintf("Failed to delete pods from node %s, %s", node.Name, err))
		ks.notifier.Error(config.EventDrain, fmt.Sprintf("%s\nError:%s", nodeDetails, err))
		return err
	}

	ks.logger.Info(fmt.Sprintf("Successfully drained node %s in %s", name, time.Now().Sub(start)))
	return nil
}

// EvictPodsFromNode drains pods from a node using k8s evict api.
func (ks KillerService) EvictPodsFromNode(name string, timeout time.Duration, evictDeleteDeadline time.Duration, gracePeriodSeconds int) error {
	start := time.Now()

	node, err := ks.kubeClient.GetNode(name)

	if err != nil {
		ks.logger.Error(fmt.Sprintf("Error fetching the node %s, %s", name, err))
		return err
	}
	nodeDetails := getNodeDetails(node)

	if err := ks.makeNodeUnschedulable(node); err != nil {
		ks.logger.Error(fmt.Sprintf("Failed to cordon the node %s, %s", node.Name, err))
		ks.notifier.Error(config.EventCordon, fmt.Sprintf("%s\nError:%s", nodeDetails, err))
		return err
	}

	// evictionGracePeriod is time period that SA is allowed to evict the pods
	// if evict takes longer than evictionGracePeriod, SA treats remaining pods
	// as misconfigured and delete them instead.
	evictionGracePeriod := timeout - evictDeleteDeadline
	ks.logger.Debug(fmt.Sprintf("Evicting pods from node %s with timeout %s", timeout, evictionGracePeriod))

	// evict pods
	err = ks.kubeClient.DrainNode(node.Name, true, evictionGracePeriod, gracePeriodSeconds)
	if err == nil {
		ks.logger.Info(fmt.Sprintf("Successfully drained node %s in %s", name, time.Now().Sub(start)))
		return nil
	}

	// check for unexpected error
	if !errors.Is(err, k8s.ErrDrainNodeTimeout) {
		ks.logger.Error(fmt.Sprintf("Failed evicting pods from node %s: %s", node.Name, err))
		ks.notifier.Error(config.EventCordon, fmt.Sprintf("%s\nFailed evicting pods from node:%s", nodeDetails, err))
		return err
	}

	// eviction took too long to complete, fallback to delete the remaining pods
	ks.logger.Info(fmt.Sprintf("Timeout evicting pods from node %s, fallback to delete", node.Name))
	ks.notifier.Info(config.EventCordon, fmt.Sprintf("%s\nTimeout evicting pods from node, fallback to delete", nodeDetails))

	err = ks.kubeClient.DrainNode(node.Name, false, evictDeleteDeadline, gracePeriodSeconds)
	if err != nil {
		ks.logger.Error(fmt.Sprintf("Failed deleting pods from node %s: %s", node.Name, err))
		ks.notifier.Error(config.EventCordon, fmt.Sprintf("%s\nFailed deleting pods from node:%s", nodeDetails, err))
		return err
	}

	ks.logger.Info(fmt.Sprintf("Successfully drained node %s in %s", name, time.Now().Sub(start)))
	return nil
}
