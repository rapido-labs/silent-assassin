package spotter

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/go-intervals/timespanset"
	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/k8s"
	"github.com/roppenlabs/silent-assassin/pkg/logger"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// whitelistStartPrefix in `YYYY-MM-DDT` format, can be anthing
	whitelistStartPrefix = "2000-01-01T"

	// whitelistEndPrefix in `YYYY-MM-DDT` format, has to be whitelistStartPrefix plus one day
	whitelistEndPrefix = "2000-01-02T"

	// whitelistTimePostfix in `:ssZ` format, can be anything
	whitelistTimePostfix = ":00Z"
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
		creationTimeStamp := node.GetCreationTimestamp()
		if _, ok := nodeAnnotations[config.SpotterExpiryTimeAnnotation]; ok {
			continue
		}
		if nodeAnnotations == nil {
			nodeAnnotations = make(map[string]string, 0)
		}
		expiryTime := getExpiryTimestamp(creationTimeStamp)
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

func (ss *spotterService) initWhitelist() {

	wlStr := ss.cp.GetStringSlice(config.SpotterWhiteListIntervalHours)
	for _, wl := range wlStr {
		times := strings.Split(wl, "-")
		start, err := time.Parse(time.RFC3339, whitelistStartPrefix+times[0]+whitelistTimePostfix)
		if err != nil {
			ss.logger.Error(fmt.Sprintf("Error parsing WhiteList Start date Reason: %v", err))
			panic(err)
		}
		end, err := time.Parse(time.RFC3339, whitelistStartPrefix+times[1]+whitelistTimePostfix)
		if err != nil {
			ss.logger.Error(fmt.Sprintf("Error parsing WhiteList End date Reason: %v", err))
			panic(err)
		}
		ss.whiteListIntervals.Insert(start, end)
	}
}

func getExpiryTimestamp(creationTs v1.Time) string {

	return creationTs.Add(time.Hour * 12).Format(time.RFC1123Z)
}
