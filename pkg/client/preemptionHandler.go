package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/roppenlabs/silent-assassin/internal/restclient"
	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/logger"
	"github.com/roppenlabs/silent-assassin/pkg/metadataclient"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	maintainanceEventSuffix = "instance/maintenance-event"
	preemptionEventSuffix   = "instance/preempted"

	maintenanceEventTerminate = "TERMINATE_ON_HOST_MAINTENANCE"
	preemptionEvent           = "TRUE"
)

type PreemptionNotifierService struct {
	logger             logger.IZapLogger
	pendingTermination chan bool
	metadata           metadataclient.IMetadata
	httpClient         restclient.IHTTPClient
	cp                 config.IProvider
}

//NewPreemptionHandler creates an instance of preemption
func NewPreemptionNotificationService(logger logger.IZapLogger, cp config.IProvider) PreemptionNotifierService {
	httpClient := http.DefaultClient
	return PreemptionNotifierService{
		pendingTermination: make(chan bool),
		metadata:           metadataclient.Mclient{},
		httpClient:         httpClient,
		cp:                 cp,
	}
}

func (pns *PreemptionNotifierService) handleTermination(state string, exists bool) error {
	if !exists {
		pns.logger.Error("Preemption event metadata API deleted unexpectedly")
	}

	if state == preemptionEvent || state == maintenanceEventTerminate {
		pns.pendingTermination <- true
	}

	return nil
}

func (pns *PreemptionNotifierService) watch() <-chan bool {
	go wait.Forever(func() {
		err := pns.metadata.Subscribe(preemptionEventSuffix, pns.handleTermination)

		if err != nil {
			pns.logger.Error(fmt.Sprintf("Failed to get preemption status - %s", err.Error()))
		}

	}, time.Second)

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
func (pns *PreemptionNotifierService) reuestGracefullDeleteionOfPods(nodeName string) {
	pns.logger.Debug("Calling Server to drain the node")
	data, err := json.Marshal(preemptionRequest{nodeName})
	if err != nil {
		pns.logger.Error(fmt.Sprintf("Error building request %s", err))
	}
	b := bytes.NewBuffer(data)

	req, err := http.NewRequest("POST", pns.cp.GetString(config.ServerHost), b)
	if err != nil {
		panic(err.Error())
	}

	req.Header.Set("Content-type", "application/json")
	var res *http.Response
	for i := 0; i < pns.cp.GetInt(config.ClientServerRetries); i++ {
		res, err = pns.httpClient.Do(req)
		if err != nil || res.StatusCode != 204 {
			pns.logger.Error(fmt.Sprintf("Trial: %d Error calling Server: %s", i+1, err.Error()))
			continue
		}
		break
	}
	if res != nil {
		res.Body.Close()
	}
}

func (pns *PreemptionNotifierService) Start(ctx context.Context, wg *sync.WaitGroup) {
	nodeName, err := pns.metadata.InstanceName()
	pns.logger.Info(fmt.Sprintf("Nodename is %s", nodeName))
	if err != nil {
		pns.logger.Error("Failed to fetch nodeName from metadata server")
		panic(err.Error())
	}
	for {
		select {
		case <-ctx.Done():
			pns.logger.Info("Shutting down client")
			wg.Done()
			return
		case termination := <-pns.watch():
			if termination {
				pns.reuestGracefullDeleteionOfPods(nodeName)
			}
		}
	}
}
