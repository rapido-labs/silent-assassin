package spotter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/go-intervals/timespanset"
	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/k8s"
	"github.com/roppenlabs/silent-assassin/pkg/logger"
	"github.com/roppenlabs/silent-assassin/pkg/notifier"
	v1 "k8s.io/api/core/v1"
)

type spotterService struct {
	cp                 config.IProvider
	logger             logger.IZapLogger
	kubeClient         k8s.IKubernetesClient
	whiteListIntervals *timespanset.Set
	notifier           notifier.INotifierClient
}

func NewSpotterService(cp config.IProvider, zl logger.IZapLogger, kc k8s.IKubernetesClient, nf notifier.INotifierClient) spotterService {
	return spotterService{
		cp:         cp,
		logger:     zl,
		kubeClient: kc,
		notifier:   nf,
	}
}

func (ss spotterService) Start(ctx context.Context, wg *sync.WaitGroup) {

	ss.logger.Info(fmt.Sprintf("Starting Spotter Loop - Poll Interval : %d", ss.cp.GetInt(config.SpotterPollIntervalMs)))

	ss.initWhitelist()

	for {
		select {
		case <-ctx.Done():
			ss.logger.Info("Shutting down spotter service")
			wg.Done()
			return
		default:
			ss.spot()
			time.Sleep(time.Millisecond * time.Duration(ss.cp.GetInt(config.SpotterPollIntervalMs)))
		}
	}
}

func getNodeDetails(node v1.Node) string {
	return fmt.Sprintf("Node: %s\n"+
		"Creation Time: %s\n"+
		"ExpiryTime: %s",
		node.Name, node.CreationTimestamp, node.Annotations[config.ExpiryTimeAnnotation])
}

func (ss spotterService) spot() {
	nodes, err := ss.kubeClient.GetNodes(ss.cp.GetString(config.NodeSelectors))

	if err != nil {
		ss.logger.Error(fmt.Sprintf("Error getting nodes %s", err.Error()))
		ss.notifier.Error(config.EventGetNodes, fmt.Sprintf("Error getting nodes %s", err.Error()))
		return
	}
	ss.logger.Debug(fmt.Sprintf("Fetched %d node(s)", len(nodes.Items)))

	for _, node := range nodes.Items {
		nodeAnnotations := node.GetAnnotations()

		if _, ok := nodeAnnotations[config.ExpiryTimeAnnotation]; ok {
			continue
		}
		if nodeAnnotations == nil {
			nodeAnnotations = make(map[string]string, 0)
		}
		expiryTime := ss.getExpiryTimestamp(node)
		ss.logger.Debug(fmt.Sprintf("spot() : Node = %v Creation Time = [ %v ] Expirty Time [ %v ]", node.Name, node.GetCreationTimestamp(), expiryTime))
		nodeAnnotations[config.ExpiryTimeAnnotation] = expiryTime

		node.SetAnnotations(nodeAnnotations)
		err := ss.kubeClient.UpdateNode(node)
		nodeDetails := getNodeDetails(node)

		if err != nil {
			ss.logger.Error(fmt.Sprintf("Failed to annotate node : %s", node.ObjectMeta.Name))
			ss.notifier.Error(config.EventAnnotate, fmt.Sprintf("%s\nError:%s", nodeDetails, err.Error()))
			continue
		}

		ss.logger.Info(fmt.Sprintf("Annotated node : %s", node.ObjectMeta.Name))
		ss.notifier.Info(config.EventAnnotate, nodeDetails)

	}

}
