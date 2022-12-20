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
	var nodeTerminationRequest NodeTerminationRequest
	if err := json.NewDecoder(r.Body).Decode(&nodeTerminationRequest); err != nil {
		s.logger.Error(fmt.Sprintf("Error decoding the request body %s", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	node, err := s.killer.GetNode(nodeTerminationRequest.Name)

	if err != nil {
		s.logger.Error(fmt.Sprintf("Error fetching the node %s, %s", nodeTerminationRequest.Name, err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	nodePool := node.Labels[s.cp.GetString(config.NodePoolLabel)]

	nodesPreempted.WithLabelValues(nodePool).Inc()
	err = s.killer.EvacuatePodsFromNode(nodeTerminationRequest.Name, s.cp.GetUint32(config.KillerDrainingTimeoutWhenNodePreemptedMs), true)

	if err != nil {
		s.logger.Error(fmt.Sprintf("Error evacuating pods from node %s", node.Name))
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
}
