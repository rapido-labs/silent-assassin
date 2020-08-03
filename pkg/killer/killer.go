package killer

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/gcloud"
	"github.com/roppenlabs/silent-assassin/pkg/k8s"
	"github.com/roppenlabs/silent-assassin/pkg/logger"
	"github.com/roppenlabs/silent-assassin/pkg/notifier"
	v1 "k8s.io/api/core/v1"
)

type killerService struct {
	cp           config.IProvider
	logger       logger.IZapLogger
	kubeClient   k8s.IKubernetesClient
	gcloudClient gcloud.IGCloudClient
	notifier     notifier.INotifierClient
}

func NewKillerService(cp config.IProvider, zl logger.IZapLogger, kc k8s.IKubernetesClient, gc gcloud.IGCloudClient, nf notifier.INotifierClient) killerService {
	return killerService{
		cp:           cp,
		logger:       zl,
		kubeClient:   kc,
		gcloudClient: gc,
		notifier:     nf,
	}
}

func (ks killerService) Start(ctx context.Context, wg *sync.WaitGroup) {
	ks.logger.Info(fmt.Sprintf("Starting Killer Loop - Poll Interval : %d", ks.cp.GetInt(config.KillerPollIntervalMs)))

	for {
		select {
		case <-ctx.Done():
			ks.logger.Info("Shutting down killer service")
			wg.Done()
			return
		default:
			ks.kill()
			time.Sleep(time.Millisecond * time.Duration(ks.cp.GetInt(config.KillerPollIntervalMs)))
		}
	}
}

//getExpiryTime returns the expiry time set on the node
//using the annotation silent-assassin/expiry-time
func getExpiryTime(node v1.Node) string {

	if timestamp, ok := node.ObjectMeta.Annotations[config.SpotterExpiryTimeAnnotation]; ok {
		return timestamp
	}

	return ""
}

//findExpiredTimeNodes gets the list of nodes whose expiry time set is older than current time
// This eligible for deletion.
func (ks killerService) findExpiredTimeNodes(labels []string) []v1.Node {
	nodeList := ks.kubeClient.GetNodes(labels)
	nodesToBeDeleted := make([]v1.Node, 0)

	now := time.Now().UTC()

	for _, node := range nodeList.Items {
		timestamp := getExpiryTime(node)
		if timestamp == "" {
			ks.logger.Warn(fmt.Sprintf("Node %s does not have %s annotation set", node.Name, config.SpotterExpiryTimeAnnotation))
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
func (ks killerService) makeNodeUnschedulable(node v1.Node) error {

	node.Spec.Unschedulable = true
	err := ks.kubeClient.UpdateNode(node)
	return err

}

//filterOutPodByOwnerReferenceKind clears the pods with ownerKind specified in the second argument,
//from the slice of pods specified in the first argument.
func filterOutPodByOwnerReferenceKind(podList []v1.Pod, kind string) (output []v1.Pod) {
	for _, pod := range podList {
		for _, ownerReference := range pod.ObjectMeta.OwnerReferences {
			if ownerReference.Kind != kind {
				output = append(output, pod)
			}
		}
	}

	return output
}

//getPodsToBeDeleted gets the list od pods that needs to be deleted before
//deleting the k8s node.
func (ks killerService) getPodsToBeDeleted(name string) ([]v1.Pod, error) {
	podList, err := ks.kubeClient.GetPodsInNode(name)

	if err != nil {
		return podList, err
	}
	// Filter out DaemonSet from the list of pods
	filteredPodList := filterOutPodByOwnerReferenceKind(podList, "DaemonSet")
	return filteredPodList, err
}

//waitforDrainToFinish function waits for the pods that were deleted by startDrainNode method
//to get evicted from the node. This takes  timeout as an argument, if the draining of nodes
//takes more time than the specified timeout, then the function returns timeout error.
func (ks killerService) waitforDrainToFinish(nodeName string, timeout int) error {
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
		elapsed := int(time.Since(start).Milliseconds())
		if elapsed >= timeout {
			return fmt.Errorf("Drainout timed out. Drain duration exceeded %d mill seconds", timeout)
		}
	}
}

//startDrainNode will delete the pods running on the node passed
//in the arugment.
func (ks killerService) startDrainNode(nodeName string) error {
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

func (ks killerService) kill() {

	nodesToDelete := ks.findExpiredTimeNodes(ks.cp.SplitStringToSlice(config.SpotterNodeSelectors, config.CommaSeparater))

	ks.logger.Debug(fmt.Sprintf("Number of nodes to kill %d", len(nodesToDelete)))
	for _, node := range nodesToDelete {
		ks.logger.Info(fmt.Sprintf("Processing node %s", node.Name))
		nodeDetail := fmt.Sprintf("Node: %s\nCreation Time: %s\nExpiryTime: %s", node.Name, node.CreationTimestamp, node.Annotations[config.SpotterExpiryTimeAnnotation])

		//Cordone the node. This will make this node unschedulable for new pods.
		if err := ks.makeNodeUnschedulable(node); err != nil {
			ks.logger.Error(fmt.Sprintf("Failed to cordon the node %s, %s", node.Name, err.Error()))
			continue
		}
		ks.logger.Info(fmt.Sprintf("Cordoned node %s", node.Name))

		// Drain the node,
		if err := ks.startDrainNode(node.Name); err != nil {
			ks.logger.Error(fmt.Sprintf("Failed to drain the node %s, %s", node.Name, err.Error()))
			ks.notifier.Error("DRAIN", fmt.Sprintf("%s\nError:%s", nodeDetail, err.Error()))
			continue
		}
		ks.notifier.Info("DRAIN", nodeDetail)

		//Wait for the pods to get evicted.
		drainingTimeout := ks.cp.GetInt(config.KillerDrainingTimeoutMs)

		if err := ks.waitforDrainToFinish(node.Name, drainingTimeout); err != nil {
			ks.logger.Error(fmt.Sprintf("Error while waiting for drain on node %s, %s", node.Name, err.Error()))
			ks.notifier.Error("DRAIN", fmt.Sprintf("%s\nError:%s", nodeDetail, err.Error()))
			continue
		}
		ks.logger.Info(fmt.Sprintf("Successfully drained the node %s", node.Name))

		// Delete the k8s node
		ks.logger.Info(fmt.Sprintf("Deleting node %s", node.Name))
		if err := ks.kubeClient.DeleteNode(node.Name); err != nil {
			ks.logger.Info(fmt.Sprintf("Error deleting the node %s", node.Name))
			ks.notifier.Error("DELETE_NODE", fmt.Sprintf("%s\nError:%s", nodeDetail, err.Error()))
			continue
		}
		ks.notifier.Info("DELETE NODE", nodeDetail)

		// Delete gcloud instance.
		projectID, zone := getProjectIDAndZoneFromNode(node)
		ks.logger.Info(fmt.Sprintf("Deletig google instance %s", node.Name))
		if err := ks.gcloudClient.DeleteNode(projectID, zone, node.Name); err != nil {
			ks.logger.Error(fmt.Sprintf("Could not kill the node %s", node.Name))
			ks.notifier.Error("DELETE_INSTANCE", fmt.Sprintf("%s\nError:%s", nodeDetail, err.Error()))
			continue
		}
		ks.notifier.Info("DELETE_INSTANCE", nodeDetail)
	}
}
