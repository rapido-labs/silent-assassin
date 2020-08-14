package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/roppenlabs/silent-assassin/pkg/config"
)

type Node struct {
	Name string
}

//handlePreemption handles POST request on /termination. This deletes the pods on the node requested.
func (s Server) handleTermination(w http.ResponseWriter, r *http.Request) {
	var node Node

	if err := json.NewDecoder(r.Body).Decode(&node); err != nil {
		s.logger.Error(fmt.Sprintf("Error decoding the request body %s", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := s.killer.EvacuatePodsFromNode(node.Name, s.cp.GetUint32(config.KillerDrainingTimeoutWhenNodePreemptedMs), true)

	if err != nil {
		s.logger.Error(fmt.Sprintf("Error evacuating pods from node %s"))
	}
}
