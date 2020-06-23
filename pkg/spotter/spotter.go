package spotter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/k8s"
	"github.com/roppenlabs/silent-assassin/pkg/logger"
)

type spotterService struct {
	cp         config.IProvider
	logger     logger.IZapLogger
	kubeClient k8s.IKubernetesClient
}

func NewSpotterService(cp config.IProvider, zl logger.IZapLogger, kc k8s.IKubernetesClient) spotterService {
	return spotterService{
		cp:         cp,
		logger:     zl,
		kubeClient: kc,
	}
}

func (ss spotterService) Start(ctx context.Context, wg *sync.WaitGroup) {
	ss.logger.Info(fmt.Sprintf("Starting Spotter Loop with a delay interval of %d", ss.cp.GetInt(config.SpotterPollIntervalMs)))

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
		creationTimeStamp := node.GetCreationTimestamp()
		if _, ok := nodeAnnotations[config.SpotterExpiryTimeAnnotation]; ok {
			continue
		}
		if nodeAnnotations == nil {
			nodeAnnotations = make(map[string]string, 0)
		}
		expiryTime := creationTimeStamp.Add(time.Hour * 12).Format(time.RFC1123Z)
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
