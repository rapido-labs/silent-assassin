package killer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/gcloud"
	"github.com/roppenlabs/silent-assassin/pkg/k8s"
	"github.com/roppenlabs/silent-assassin/pkg/logger"
	"github.com/roppenlabs/silent-assassin/pkg/notifier"
)

var (
	nodesKilled = promauto.NewCounter(prometheus.CounterOpts{
		Name: "nodes_killed",
		Help: "The total number of nodes killed when they reach their expiry time",
	})
)

type KillerService struct {
	cp           config.IProvider
	logger       logger.IZapLogger
	kubeClient   k8s.IKubernetesClient
	gcloudClient gcloud.IGCloudClient
	notifier     notifier.INotifierClient
}

func NewKillerService(cp config.IProvider, zl logger.IZapLogger, kc k8s.IKubernetesClient, gc gcloud.IGCloudClient, nf notifier.INotifierClient) KillerService {
	return KillerService{
		cp:           cp,
		logger:       zl,
		kubeClient:   kc,
		gcloudClient: gc,
		notifier:     nf,
	}
}

func (ks KillerService) Start(ctx context.Context, wg *sync.WaitGroup) {
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

func (ks KillerService) kill() {

	nodesToDelete := ks.findExpiredTimeNodes(ks.cp.GetString(config.NodeSelectors))

	ks.logger.Debug(fmt.Sprintf("Number of nodes to kill %d", len(nodesToDelete)))
	for _, node := range nodesToDelete {
		ks.logger.Info(fmt.Sprintf("Processing node %s", node.Name))

		nodesKilled.Inc()
		if err := ks.EvacuatePodsFromNode(node.Name, ks.cp.GetUint32(config.KillerDrainingTimeoutWhenNodeExpiredMs), false); err != nil {
			ks.logger.Error(fmt.Sprintf("Error evacuating node:%s, %s", node.Name, err.Error()))
			continue
		}

		ks.deleteNode(node)

	}
}
