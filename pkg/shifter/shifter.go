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

//Suggest a better name
type npShiftConf struct {
	destNodePool       string
	sourceMinNodeCount int64
}

type ShifterService struct {
	cp                 config.IProvider
	logger             logger.IZapLogger
	kubeClient         k8s.IKubernetesClient
	gcloudClient       gcloud.IGCloudClient
	killer             killer.KillerService
	notifier           notifier.INotifierClient
	nodePoolMap        map[string]npShiftConf
	whiteListIntervals []wlInterval
}

func NewShifterService(cp config.IProvider, zl logger.IZapLogger, kc k8s.IKubernetesClient, gc gcloud.IGCloudClient, nf notifier.INotifierClient, kl killer.KillerService) ShifterService {
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
				if ss.timeWithinWLIntervalCheck(interval.start, interval.end, now) {
					ss.shift()
				}
			}
			time.Sleep(time.Millisecond * time.Duration(ss.cp.GetInt(config.ShifterPollIntervalMs)))
		}
	}
}

//getNodePoolMap finds out fallback on-demand nodePools and their respective preemptible node-pools.
func (ss *ShifterService) getNodePoolMap() error {

	nps, err := ss.gcloudClient.ListNodePools()

	ss.logger.Info(fmt.Sprintf("Nodepools in the cluster %v", nps))

	if err != nil {
		return err
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
	ss.nodePoolMap = make(map[string]npShiftConf)
	for _, pnp := range preemptibleNodePools {
		for _, onp := range onDemandNodePools {
			if reflect.DeepEqual(pnp.Config.Labels, onp.Config.Labels) && pnp.Config.MachineType == onp.Config.MachineType {

				ss.nodePoolMap[onp.Name] = npShiftConf{
					destNodePool:       pnp.Name,
					sourceMinNodeCount: onp.Autoscaling.MinNodeCount,
				}
			}
		}
	}
	return nil
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

		node.Spec.Unschedulable = true
		err := ss.kubeClient.UpdateNode(node)
		return err
	}
	return nil
}

func (ss ShifterService) shift() {

	//Create a nodepool map to determine source fallback on-demand nodepool and
	//their respective preemptible preemptible nodepools.
	err := ss.getNodePoolMap()
	if err != nil {
		ss.logger.Error(fmt.Sprintf("Error creating the noedpool map: %v", err.Error()))
		ss.notifier.Error(config.EventGetNodes, fmt.Sprintf("Error creating the nodepool %v", err.Error()))
		return
	}

	// Iterate through each fallback ond-demand nodepool to see if they have node-pool size
	// greater than the min node count.
	for sourceNodePool, npInfo := range ss.nodePoolMap {

		sourceNPSelector := fmt.Sprintf("cloud.google.com/gke-nodepool=%s", sourceNodePool)
		sourceNPSize, err := ss.getNodePoolSize(sourceNPSelector)
		if err != nil {
			ss.notifier.Error(config.EventGetNodes, fmt.Sprintf("Error getting node size of nodepool %v, %v", sourceNodePool, err.Error()))
			ss.logger.Error(fmt.Sprintf("Error getting node size of nodepool %v", sourceNodePool))
			continue
		}

		destSelector := fmt.Sprintf("cloud.google.com/gke-nodepool=%s", npInfo.destNodePool)

		//Shift the nodes when size of fallback on-demand nodepool is greater than preemptible node-pool.
		if sourceNPSize > npInfo.sourceMinNodeCount {

			sourceNodes, err := ss.kubeClient.GetNodes(sourceNPSelector)

			if err != nil {
				ss.notifier.Error(config.EventGetNodes, fmt.Sprintf("Error getting nodes in %v nodepool, %v", sourceNodePool, err.Error()))
				ss.logger.Error(fmt.Sprintf("Error getting nodes in %v nodepool, %v", sourceNodePool, err.Error()))
				continue
			}

			destNPSize, err := ss.getNodePoolSize(destSelector)
			if err != nil {
				ss.logger.Error(fmt.Sprintf("Error fetching destination nodepool size %v\n", err.Error()))
				ss.notifier.Error(config.EventGetNodes, fmt.Sprintf("Error fetching destination nodepool size %v\n", err.Error()))
			}

			// Nodes to be added is the difference between current size of source node-pool and its min node count.
			nodesTobeAdded := sourceNPSize - npInfo.sourceMinNodeCount

			// Resize the destination nodepool to sum of curent size of destination node-pool and nodesTobeAdded.
			ss.logger.Info(fmt.Sprintf("Resizing the destination nodepool: %v node-size: %d -> %d", npInfo.destNodePool, destNPSize, destNPSize+nodesTobeAdded))
			ss.notifier.Info(config.EvenetResizeNodePool, fmt.Sprintf("Resizing the destination nodepool: %v node-size: %d -> %d", npInfo.destNodePool, destNPSize, destNPSize+nodesTobeAdded))
			err = ss.gcloudClient.SetNodePoolSize(npInfo.destNodePool, destNPSize+nodesTobeAdded, ss.cp.GetInt(config.ShifterNPResizeTimeout))
			if err != nil {
				ss.logger.Error(fmt.Sprintf("Resizing the destination nodepool: %v node-size: %d -> %d failed: %v", npInfo.destNodePool, destNPSize, destNPSize+nodesTobeAdded, err.Error()))
				ss.notifier.Error(config.EvenetResizeNodePool, fmt.Sprintf("Resizing the destination nodepool: %v node-size: %d -> %d failed: %v", npInfo.destNodePool, destNPSize, destNPSize+nodesTobeAdded, err.Error()))
				// Return, as there might not be enough preemptible resources available at the data center.
				return
			}

			// Cordon all source nodes so that no deleted workload will get scheduled in the source nodes again.
			err = ss.makeNodeUnschedulable(sourceNodes.Items)
			if err != nil {
				ss.notifier.Error(config.EventCordon, fmt.Sprintf("Error cordoning node %v", err.Error()))
				ss.logger.Error(fmt.Sprintf("Error cordoning node %v", err.Error()))
			}

			// Iterate through source nodes and drain the node.
			for _, node := range sourceNodes.Items {

				err := ss.killer.EvacuatePodsFromNode(node.Name, ss.cp.GetUint32(config.KillerDrainingTimeoutWhenNodeExpiredMs), false)

				if err != nil {
					ss.logger.Error(fmt.Sprintf("Error draining the node %v: %v", node.Name, err.Error()))
					ss.notifier.Error(config.EventDrain, fmt.Sprintf("Error draining the node %v: %v", node.Name, err.Error()))
					continue
				}

				err = ss.kubeClient.DeleteNode(node.Name)
				if err != nil {
					ss.logger.Error(fmt.Sprintf("Error deleting the node %v: %v", node.Name, err.Error()))
					ss.notifier.Error(config.EventDeleteNode, fmt.Sprintf("Error deleting the node %v: %v", node.Name, err.Error()))
					continue
				}
				//Sleep after node deletion for the workloads to stabilize
				time.Sleep(time.Millisecond * time.Duration(ss.cp.GetInt32(config.ShifterSleepAfterNodeDeletionMs)))
			}
		}
	}
}
