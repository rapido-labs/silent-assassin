package informer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/gcloud"
	"github.com/roppenlabs/silent-assassin/pkg/logger"
	"github.com/roppenlabs/silent-assassin/pkg/utils"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	maintainanceEventSuffix = "instance/maintenance-event"
	preemptionEventSuffix   = "instance/preempted"

	maintenanceEventTerminate = "TERMINATE_ON_HOST_MAINTENANCE"
	preemptionEvent           = "TRUE"
)

type node struct {
	Name string
}
type InformerService struct {
	logger             logger.IZapLogger
	pendingTermination chan bool
	metadata           gcloud.IMetadata
	httpClient         utils.IHTTPClient
	cp                 config.IProvider
}

// NewInformerService creates an instance of preemptionNotifierService
func NewInformerService(logger logger.IZapLogger, cp config.IProvider) InformerService {
	httpClient := http.DefaultClient
	return InformerService{
		logger:             logger,
		pendingTermination: make(chan bool),
		metadata:           gcloud.Mclient{},
		httpClient:         httpClient,
		cp:                 cp,
	}
}

func (pns InformerService) handleTermination(state string, exists bool) error {
	if !exists {
		pns.logger.Error("Preemption event metadata API deleted unexpectedly")
	}

	if state == preemptionEvent || state == maintenanceEventTerminate {
		pns.pendingTermination <- true
	}

	return nil
}

func (pns InformerService) watch() <-chan bool {
	//Watch for preemption event
	go wait.Forever(func() {
		err := pns.metadata.Subscribe(preemptionEventSuffix, pns.handleTermination)

		if err != nil {
			pns.logger.Error(fmt.Sprintf("Failed to get preemption status - %s", err.Error()))
		}

	}, time.Second)

	//Watch for maintainance event
	if pns.cp.GetBool(config.ClientWatchMaintainanceEvents) {
		go wait.Forever(func() {
			err := pns.metadata.Subscribe(maintainanceEventSuffix, pns.handleTermination)

			if err != nil {
				pns.logger.Error(fmt.Sprintf("Failed to get maintenance status - %s", err.Error()))
			}

		}, time.Second)
	}

	return pns.pendingTermination
}

//Request silent-assassin server of the preemption for graceful deleteion of pods in the node
func (pns InformerService) informPreemption(nodeName string) {
	pns.logger.Info(fmt.Sprintf("Calling Server to drain the node %s", nodeName))

	node := node{
		Name: nodeName,
	}
	data, err := json.Marshal(node)
	if err != nil {
		pns.logger.Error(fmt.Sprintf("Error building request %s", err))
	}

	b := bytes.NewBuffer(data)

	preemptionURI := fmt.Sprintf("%s%s", pns.cp.GetString(config.ServerHost), config.EvacuatePodsURI)
	req, err := http.NewRequest(http.MethodPost, preemptionURI, b)
	if err != nil {
		panic(err.Error())
	}

	req.Header.Set("Content-type", "application/json")

	var res *http.Response

	for i := 0; i < pns.cp.GetInt(config.ClientServerRetries); i++ {
		res, err = pns.httpClient.Do(req)
		if err != nil {
			pns.logger.Error(fmt.Sprintf("Trial %d: Error calling Server: %v", i+1, err))
			continue
		}
		if res.StatusCode != http.StatusNoContent {
			pns.logger.Error(fmt.Sprintf("Trial %d: Error calling Server response status %d", i+1, res.StatusCode))
			continue
		}
		break
	}
	if res != nil {
		res.Body.Close()
	}
}

//Start starts the preemptionNotificationService service
func (pns InformerService) Start(ctx context.Context, wg *sync.WaitGroup) {
	nodeName, err := pns.metadata.InstanceName()
	if err != nil {
		pns.logger.Error(fmt.Sprintf("Failed to fetch node name from metadata server %s", err.Error()))
		panic(err.Error())
	}
	pns.logger.Info(fmt.Sprintf("Node %s", nodeName))
	for {
		select {
		case <-ctx.Done():
			pns.logger.Info("Shutting down Client")
			wg.Done()
			return
		case termination := <-pns.watch():
			if termination {
				pns.informPreemption(nodeName)
				return
			}
		}
	}
}
