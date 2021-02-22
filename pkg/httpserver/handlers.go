package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/roppenlabs/silent-assassin/pkg/config"
)

type NodeTerminationRequest struct {
	Name string
}

//handlePreemption handles POST request on EvacuatePodsURI. This deletes the pods on the node requested.
func (s Server) handleTermination(w http.ResponseWriter, r *http.Request) {
	var node NodeTerminationRequest
	if err := json.NewDecoder(r.Body).Decode(&node); err != nil {
		s.logger.Error(fmt.Sprintf("Error decoding the request body %s", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	nodesPreempted.Inc()
	err := s.killer.EvacuatePodsFromNode(node.Name, s.cp.GetDuration(config.KillerDrainingTimeoutWhenNodePreempted), true)

	if err != nil {
		s.logger.Error(fmt.Sprintf("Error evacuating pods from node %s", node.Name))
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
}
