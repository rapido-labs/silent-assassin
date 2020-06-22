package spotter

import (
	"fmt"
	"time"

	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/k8s"
	"github.com/roppenlabs/silent-assassin/pkg/logger"
)

func Start(cp config.IProvider, zl logger.IZapLogger, kc k8s.IKubernetesClient) {

	zl.Debug(fmt.Sprintf("Starting Spotter Loop with a delay interval of %d", cp.GetInt(config.SpotterPollIntervalMs)), "START", config.LogComponentName)

	for {
		spotNodesForRemoval(cp, zl, kc)
		time.Sleep(time.Millisecond * time.Duration(cp.GetInt(config.SpotterPollIntervalMs)))
	}

}

func spotNodesForRemoval(cp config.IProvider, zl logger.IZapLogger, kc k8s.IKubernetesClient) {
	nodes := kc.GetNodes(cp.GetStringSlice(config.SpotterNodeSelectors))

	zl.Info(fmt.Sprintf("Fetched %d node(s)", len(nodes.Items)), "GET_NODES", config.LogComponentName)

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
		err := kc.AnnotateNode(node)
		if err != nil {
			zl.Error(fmt.Sprintf("Failed to annotate node : %s", node.ObjectMeta.Name), "ANNOTATE_NODES", config.LogComponentName)
			panic(err)
		}
		zl.Info(fmt.Sprintf("Annotated node : %s", node.ObjectMeta.Name), "ANNOTATE_NODES", config.LogComponentName)
	}
}
