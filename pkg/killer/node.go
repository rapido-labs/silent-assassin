package killer

import (
	"fmt"
	"strings"
	"time"

	"github.com/roppenlabs/silent-assassin/pkg/config"
	v1 "k8s.io/api/core/v1"
)

//getExpiryTime returns the expiry time set on the node
//using the annotation silent-assassin/expiry-time
func getExpiryTime(node v1.Node) string {

	if timestamp, ok := node.ObjectMeta.Annotations[config.ExpiryTimeAnnotation]; ok {
		return timestamp
	}

	return ""
}

//findExpiredTimeNodes gets the list of nodes whose expiry time set is older than current time
//These nodes are eligible for deletion.
func (ks KillerService) findExpiredTimeNodes(labelSelector string) []v1.Node {
	nodeList := ks.kubeClient.GetNodes(labelSelector)
	nodesToBeDeleted := make([]v1.Node, 0)

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

	return nodesToBeDeleted
}

//makeNodeUnschedulable function cordons the node thus disabling scheduling of
//any new pods on this node during draining.
func (ks KillerService) makeNodeUnschedulable(node v1.Node) error {

	node.Spec.Unschedulable = true
	err := ks.kubeClient.UpdateNode(node)
	return err

}

//waitforDrainToFinish function waits for the pods that were deleted by startDrainNode method
//to get evicted from the node. This takes  timeout as an argument, if the draining of nodes
//takes more time than the specified timeout, then the function returns timeout error.
func (ks KillerService) waitforDrainToFinish(nodeName string, timeout uint32) error {
	start := time.Now()
	for {
		podsPending, err := ks.getPodsToBeDeleted(nodeName)
		if err != nil {
			ks.logger.Error(fmt.Sprintf("Error fetching pods: %s", err.Error()))
			return err
		}

		if len(podsPending) == 0 {
			return nil
		}
		elapsed := uint32(time.Since(start).Milliseconds())
		if elapsed >= timeout {
			return fmt.Errorf("Drainout timed out. Drain duration exceeded %d mill seconds", timeout)
		}
	}
}

//startDrainNode will delete the pods running on the node passed
//in the arugment.
func (ks KillerService) startNodeDrain(nodeName string) error {
	filteredPodList, err := ks.getPodsToBeDeleted(nodeName)
	if err != nil {
		return err
	}
	for _, pod := range filteredPodList {
		ks.logger.Info(fmt.Sprintf("Deleting pod %s on node %s", pod.Name, nodeName))

		if err := ks.kubeClient.DeletePod(pod.Name, pod.Namespace); err != nil {
			ks.logger.Error(
				fmt.Sprintf("Error deleting the pod %s on node %s in %s namespace:%s",
					pod.Name, nodeName, pod.Namespace, err.Error(),
				),
			)
			return err
		}
	}
	return nil
}

//getProjectIDAndZoneFromNode extracts the GCP projectID and zone
//from the given node.
func getProjectIDAndZoneFromNode(node v1.Node) (string, string) {
	s := strings.Split(node.Spec.ProviderID, "/")
	projectID := s[2]
	zone := s[3]

	return projectID, zone
}

func (ks KillerService) deleteNode(node v1.Node) {
	nodeDetails := getNodeDetails(node, true)

	// Delete the k8s node
	ks.logger.Info(fmt.Sprintf("Deleting node %s", node.Name))
	if err := ks.kubeClient.DeleteNode(node.Name); err != nil {
		ks.logger.Info(fmt.Sprintf("Error deleting the node %s", node.Name))
		ks.notifier.Error(config.EventDeleteNode, fmt.Sprintf("%s\nError:%s", nodeDetails, err.Error()))
		return
	}

	ks.notifier.Info(config.EventDeleteNode, nodeDetails)

	// Delete gcloud instance.
	projectID, zone := getProjectIDAndZoneFromNode(node)
	ks.logger.Info(fmt.Sprintf("Deletig google instance %s", node.Name))
	if err := ks.gcloudClient.DeleteInstance(projectID, zone, node.Name); err != nil {
		ks.logger.Error(fmt.Sprintf("Could not kill the node %s", node.Name))
		ks.notifier.Error(config.EventDeleteInstance, fmt.Sprintf("%s\nError:%s", nodeDetails, err.Error()))
		return
	}
	ks.notifier.Info(config.EventDeleteInstance, nodeDetails)
}

func getNodeDetails(node v1.Node, preemption bool) string {
	return fmt.Sprintf("Node: %s\n"+
		"Preemption: %t\n"+
		"Creation Time: %s\n"+
		"ExpiryTime: %s",
		node.Name, preemption, node.CreationTimestamp, node.Annotations[config.ExpiryTimeAnnotation])
}

//handlePreemption handles POST request on EvacuatePodsURI. This deletes the pods on the node requested.
func (ks KillerService) EvacuatePodsFromNode(name string, timeout uint32, preemption bool) error {
	start := time.Now()

	node, err := ks.kubeClient.GetNode(name)

	if err != nil {
		ks.logger.Error(fmt.Sprintf("Error fetching the node %s, %s", name, err.Error()))
		return err
	}
	nodeDetails := getNodeDetails(node, preemption)

	if err := ks.makeNodeUnschedulable(node); err != nil {
		ks.logger.Error(fmt.Sprintf("Failed to cordon the node %s, %s", node.Name, err.Error()))
		ks.notifier.Error(config.EventCordon, fmt.Sprintf("%s\nError:%s", nodeDetails, err.Error()))
		return err
	}

	if err := ks.startNodeDrain(node.Name); err != nil {
		ks.logger.Error(fmt.Sprintf("Failed to drain the node %s, %s", node.Name, err.Error()))
		ks.notifier.Error(config.EventDrain, fmt.Sprintf("%s\nError:%s", nodeDetails, err.Error()))
		return err
	}

	if err := ks.waitforDrainToFinish(node.Name, timeout); err != nil {
		ks.logger.Error(fmt.Sprintf("Error while waiting for drain on node %s, %s", node.Name, err.Error()))
		ks.notifier.Error(config.EventDrain, fmt.Sprintf("%s\nError:%s", nodeDetails, err.Error()))
		return err
	}
	ks.logger.Info(fmt.Sprintf("Successfully drained the node %s", node.Name))
	ks.notifier.Info(config.EventDrain, nodeDetails)

	end := time.Now()
	timeTakenToEvacuatePods := end.Sub(start).Seconds()

	ks.logger.Info(fmt.Sprintf("Took %f seconds to drain the node %s", timeTakenToEvacuatePods, node.Name))
	ks.logger.Info(fmt.Sprintf("Successfully drained the node %s", node.Name))
	return nil
}
