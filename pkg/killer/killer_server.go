package killer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/roppenlabs/silent-assassin/pkg/config"
)

type preemptNode struct {
	NodeName string
}

//StartServer starts the killer server
func (ks killerService) StartServer(ctx context.Context, wg *sync.WaitGroup) {
	ks.logger.Info("Starting Killer server")

	router := mux.NewRouter()
	router.HandleFunc("/preemption", ks.handlePreemption).Methods("POST")

	host := fmt.Sprintf("0.0.0.0:%d", ks.cp.GetInt32(config.ServerPort))
	srv := &http.Server{
		Addr:    host,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			ks.logger.Error(fmt.Sprintf("Error starting server: %s", err.Error()))
			panic(err.Error())
		}
	}()

	<-ctx.Done()
	srv.Shutdown(ctx)
	wg.Done()
	return
}

//handlePreemption handles POST request on /preemption. This deletes the pods on the node requested.
func (ks killerService) handlePreemption(w http.ResponseWriter, r *http.Request) {
	var preemptibleNode preemptNode
	start := time.Now()
	if err := json.NewDecoder(r.Body).Decode(&preemptibleNode); err != nil {
		ks.logger.Error(fmt.Sprintf("Error decoding the request body %s", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	node, err := ks.kubeClient.GetNode(preemptibleNode.NodeName)
	if err != nil {
		ks.logger.Error(fmt.Sprintf("Error fetching the node %s, %s", preemptibleNode.NodeName, err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	nodeDetail := fmt.Sprintf("Node: %s\nPreemption: True\nCreation Time: %s\nExpiryTime: %s", node.Name, node.CreationTimestamp, node.Annotations[config.SpotterExpiryTimeAnnotation])

	if err := ks.makeNodeUnschedulable(node); err != nil {
		ks.logger.Error(fmt.Sprintf("Failed to cordon the node %s, %s", node.Name, err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		ks.notifier.Error("CORDON", fmt.Sprintf("%s\nError:%s", nodeDetail, err.Error()))
		return
	}

	if err := ks.startDrainNode(node.Name); err != nil {
		ks.logger.Error(fmt.Sprintf("Failed to drain the node %s, %s", node.Name, err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		ks.notifier.Error("DRAIN", fmt.Sprintf("%s\nError:%s", nodeDetail, err.Error()))
		return
	}

	if err := ks.waitforDrainToFinish(node.Name, ks.cp.GetInt(config.KillerPreemptionDrainTimeout)); err != nil {
		ks.logger.Error(fmt.Sprintf("Error while waiting for drain on node %s, %s", node.Name, err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		ks.notifier.Error("DRAIN", fmt.Sprintf("%s\nError:%s", nodeDetail, err.Error()))
		return
	}
	end := time.Now()
	timeTakenToDrain := end.Sub(start).Seconds()
	ks.logger.Info(fmt.Sprintf("Took %f seconds to drain the node %s", timeTakenToDrain, node.Name))
	ks.notifier.Info("DRAIN", nodeDetail)

	ks.logger.Info(fmt.Sprintf("Successfully drained the node %s", node.Name))
	w.WriteHeader(http.StatusNoContent)
}
