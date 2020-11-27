package shifter

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/gcloud"
	"github.com/roppenlabs/silent-assassin/pkg/k8s"
	"github.com/roppenlabs/silent-assassin/pkg/killer"
	"github.com/roppenlabs/silent-assassin/pkg/logger"
	"github.com/roppenlabs/silent-assassin/pkg/notifier"
	container "google.golang.org/api/container/v1"
	v1 "k8s.io/api/core/v1"
)

const timeLayout string = "15:04"

type npShiftConf struct {
	preemptibleNP        string
	onDemandMinNodeCount int64
}

type ShifterService struct {
	cp                 config.IProvider
	logger             logger.IZapLogger
	kubeClient         k8s.IKubernetesClient
	gcloudClient       gcloud.IGCloudClient
	killer             killer.IKiller
	notifier           notifier.INotifierClient
	whiteListIntervals []wlInterval
}

func NewShifterService(cp config.IProvider, zl logger.IZapLogger, kc k8s.IKubernetesClient, gc gcloud.IGCloudClient, nf notifier.INotifierClient, kl killer.IKiller) ShifterService {
	return ShifterService{
		cp:           cp,
		logger:       zl,
		kubeClient:   kc,
		gcloudClient: gc,
		notifier:     nf,
		killer:       kl,
	}
}

func (ss ShifterService) Start(ctx context.Context, wg *sync.WaitGroup) {
	ss.logger.Info(fmt.Sprintf("Starting Shifter Loop - Poll Interval: %d", ss.cp.GetInt(config.ShifterPollIntervalMs)))
	ss.initWhitelist()

	for {
		select {
		case <-ctx.Done():
			ss.logger.Info("Shutting down Shifter service")
			wg.Done()
			return
		default:
			// Check if current time is in bettween whiteListIntervals. If yes, run ss.shift().
			t := time.Now().UTC()
			hour := fmt.Sprintf("%02d", t.Hour())
			mins := fmt.Sprintf("%02d", t.Minute())
			fmtTime := fmt.Sprintf("%s:%s", hour, mins)
			now, err := time.Parse(timeLayout, fmtTime)

			if err != nil {
				panic(err)
			}

			for _, interval := range ss.whiteListIntervals {
				if timeWithinWLIntervalCheck(interval.start, interval.end, now) {
					ss.shift()
				}
			}
			time.Sleep(time.Millisecond * time.Duration(ss.cp.GetInt(config.ShifterPollIntervalMs)))
		}
	}
}

//getNodePoolMap finds out fallback on-demand nodePools and their respective preemptible node-pools.
func (ss *ShifterService) getNodePoolMap() (map[string]npShiftConf, error) {
	nodePoolMap := make(map[string]npShiftConf)
	nps, err := ss.gcloudClient.ListNodePools()

	ss.logger.Info(fmt.Sprintf("Nodepools in the cluster %v", nps))

	if err != nil {
		return nodePoolMap, err
	}

	var preemptibleNodePools, onDemandNodePools []container.NodePool

	for _, np := range nps {
		if np.Config.Preemptible {
			preemptibleNodePools = append(preemptibleNodePools, *np)
		} else {
			onDemandNodePools = append(onDemandNodePools, *np)
		}
	}

	// Loop over preemptible nodepools and find their corresponding fallback on-demand node-pool.
	for _, pnp := range preemptibleNodePools {
		for _, onp := range onDemandNodePools {
			if reflect.DeepEqual(pnp.Config.Labels, onp.Config.Labels) && pnp.Config.MachineType == onp.Config.MachineType {

				nodePoolMap[onp.Name] = npShiftConf{
					preemptibleNP:        pnp.Name,
					onDemandMinNodeCount: onp.Autoscaling.MinNodeCount,
				}
			}
		}
	}
	return nodePoolMap, nil
}

//getNodePoolSize gets the node-pool size by checking maximum number of nodes in the available zones.
//This works for single and multi zone clusters.
func (ss ShifterService) getNodePoolSize(selector string) (int64, error) {
	nodes, err := ss.kubeClient.GetNodes(selector)

	if err != nil {
		return 0, err
	}

	nodeZoneWise := make(map[string]int64)

	for _, node := range nodes.Items {
		zone := node.ObjectMeta.Labels["failure-domain.beta.kubernetes.io/zone"]
		if _, ok := nodeZoneWise[zone]; ok {
			nodeZoneWise[zone]++
		} else {
			nodeZoneWise[zone] = 1
		}
	}

	var max int64

	for _, count := range nodeZoneWise {
		if count > max {
			max = count
		}
	}

	return max, err
}

func (ss ShifterService) makeNodeUnschedulable(nodes []v1.Node) error {

	for _, node := range nodes {
		ss.logger.Info(fmt.Sprintf("Cordoning node %v", node.Name))
		recentNodeObject, err := ss.kubeClient.GetNode(node.Name)
		if err != nil {
			return err
		}
		recentNodeObject.Spec.Unschedulable = true
		err = ss.kubeClient.UpdateNode(recentNodeObject)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ss ShifterService) shift() {
	numberofZones := ss.gcloudClient.GetNumberOfZones()
	//Create a nodepool map to determine source fallback on-demand nodepool and
	//their respective preemptible preemptible nodepools.
	nodePoolMap, err := ss.getNodePoolMap()
	if err != nil {
		ss.logger.Error(fmt.Sprintf("Error creating the nodepool map: %v", err.Error()))
		ss.notifier.Error(config.EventGetNodes, fmt.Sprintf("Error creating the nodepool %v", err.Error()))
		return
	}

	// Iterate through each fallback ond-demand nodepool to see if they have node-pool size
	// greater than the min node count.
	for onDemandNodePool, npInfo := range nodePoolMap {

		onDemandNPSelector := fmt.Sprintf("cloud.google.com/gke-nodepool=%s", onDemandNodePool)
		onDemandNPSize, err := ss.getNodePoolSize(onDemandNPSelector)
		if err != nil {
			ss.notifier.Error(config.EventGetNodes, fmt.Sprintf("Error getting node size of nodepool %v, %v", onDemandNodePool, err.Error()))
			ss.logger.Error(fmt.Sprintf("Error getting node size of nodepool %v", onDemandNodePool))
			continue
		}

		preemptibleNPSelector := fmt.Sprintf("cloud.google.com/gke-nodepool=%s", npInfo.preemptibleNP)

		//Shift the nodes when size of fallback on-demand nodepool is greater than preemptible node-pool.
		if onDemandNPSize > npInfo.onDemandMinNodeCount {

			onDemandNodes, err := ss.kubeClient.GetNodes(onDemandNPSelector)

			if err != nil {
				ss.notifier.Error(config.EventGetNodes, fmt.Sprintf("Error getting nodes in %v nodepool, %v", onDemandNodePool, err.Error()))
				ss.logger.Error(fmt.Sprintf("Error getting nodes in %v nodepool, %v", onDemandNodePool, err.Error()))
				continue
			}

			// Cordon all source nodes so that no deleted workload will get scheduled in the source nodes again.
			err = ss.makeNodeUnschedulable(onDemandNodes.Items)
			if err != nil {
				ss.notifier.Error(config.EventCordon, fmt.Sprintf("Error cordoning node %v", err.Error()))
				ss.logger.Error(fmt.Sprintf("Error cordoning node %v", err.Error()))
				continue
			}
			nodesDeleted := 0
			// Iterate through source nodes and drain the node.
			for _, node := range onDemandNodes.Items {

				if nodesDeleted%numberofZones == 0 {

					preemptibleNPSize, err := ss.getNodePoolSize(preemptibleNPSelector)
					if err != nil {
						ss.logger.Error(fmt.Sprintf("Error fetching preemptible nodepool size %v\n", err.Error()))
						ss.notifier.Error(config.EventGetNodes, fmt.Sprintf("Error fetching destination nodepool size %v\n", err.Error()))
						return
					}

					// Resize the preemptible nodepool to sum of curent size of preemptible node-pool and one
					ss.logger.Info(fmt.Sprintf("Resizing the preemptible nodepool: %v node-size: %d -> %d", npInfo.preemptibleNP, preemptibleNPSize, preemptibleNPSize+1))
					ss.notifier.Info(config.EventResizeNodePool, fmt.Sprintf("Resizing the preemptible nodepool: %v node-size: %d -> %d", npInfo.preemptibleNP, preemptibleNPSize, preemptibleNPSize+1))
					err = ss.gcloudClient.SetNodePoolSize(npInfo.preemptibleNP, preemptibleNPSize+1, ss.cp.GetInt(config.ShifterNPResizeTimeout))
					if err != nil {
						ss.logger.Error(fmt.Sprintf("Resizing the preemptible nodepool: %v node-size: %d -> %d failed: %v", npInfo.preemptibleNP, preemptibleNPSize, preemptibleNPSize+1, err.Error()))
						ss.notifier.Error(config.EventResizeNodePool, fmt.Sprintf("Resizing the preemptible nodepool: %v node-size: %d -> %d failed: %v", npInfo.preemptibleNP, preemptibleNPSize, preemptibleNPSize+1, err.Error()))
						// Return, as there might not be enough preemptible resources available at the data center.
						return
					}
				}
				err := ss.killer.EvacuatePodsFromNode(node.Name, ss.cp.GetUint32(config.KillerDrainingTimeoutWhenNodeExpiredMs), false)

				if err != nil {
					ss.logger.Error(fmt.Sprintf("Error draining the node %v: %v", node.Name, err.Error()))
					ss.notifier.Error(config.EventDrain, fmt.Sprintf("Error draining the node %v: %v", node.Name, err.Error()))
					continue
				}

				ss.logger.Info(fmt.Sprintf("Deleting the node %v", node.Name))
				ss.notifier.Info(config.EventDeleteNode, fmt.Sprintf("Deleting the node %v", node.Name))
				err = ss.kubeClient.DeleteNode(node.Name)
				if err != nil {
					ss.logger.Error(fmt.Sprintf("Error deleting the node %v: %v", node.Name, err.Error()))
					ss.notifier.Error(config.EventDeleteNode, fmt.Sprintf("Error deleting the node %v: %v", node.Name, err.Error()))
					continue
				}
				nodesDeleted++

				//Sleep after node deletion for the workloads to stabilize
				time.Sleep(time.Millisecond * time.Duration(ss.cp.GetInt32(config.ShifterSleepAfterNodeDeletionMs)))
			}
		}
	}
}
