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
)

type spotterService struct {
	cp                 config.IProvider
	logger             logger.IZapLogger
	kubeClient         k8s.IKubernetesClient
	whiteListIntervals *timespanset.Set
}

func NewSpotterService(cp config.IProvider, zl logger.IZapLogger, kc k8s.IKubernetesClient) spotterService {
	return spotterService{
		cp:         cp,
		logger:     zl,
		kubeClient: kc,
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

func (ss spotterService) spot() {

	nodes := ss.kubeClient.GetNodes(ss.cp.GetStringSlice(config.SpotterNodeSelectors))

	ss.logger.Debug(fmt.Sprintf("Fetched %d node(s)", len(nodes.Items)))

	for _, node := range nodes.Items {
		nodeAnnotations := node.GetAnnotations()

		if _, ok := nodeAnnotations[config.SpotterExpiryTimeAnnotation]; ok {
			continue
		}
		if nodeAnnotations == nil {
			nodeAnnotations = make(map[string]string, 0)
		}
		expiryTime := ss.getExpiryTimestamp(node)
		ss.logger.Debug(fmt.Sprintf("spot() : Node = %v Creation Time = [ %v ] Expirty Time [ %v ]", node.Name, node.GetCreationTimestamp(), expiryTime))
		nodeAnnotations[config.SpotterExpiryTimeAnnotation] = expiryTime

		node.SetAnnotations(nodeAnnotations)
		err := ss.kubeClient.AnnotateNode(node)
		if err != nil {
			ss.logger.Error(fmt.Sprintf("Failed to annotate node : %s", node.ObjectMeta.Name))
			panic(err)
		}
		ss.logger.Info(fmt.Sprintf("Annotated node : %s", node.ObjectMeta.Name))
	}

}
