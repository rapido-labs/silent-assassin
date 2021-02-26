package k8s

import (
	"context"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"time"

	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/logger"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesClient struct {
	*kubernetes.Clientset
	ctx    context.Context
	logger logger.IZapLogger
}

type IKubernetesClient interface {
	GetNodes(labelSelector string) (*v1.NodeList, error)
	GetNode(name string) (v1.Node, error)
	GetPodsInNode(name string) ([]v1.Pod, error)
	DeleteNode(name string) error
	UpdateNode(node v1.Node) error
	DrainNode(name string, useEvict bool, timeout time.Duration, gracePeriodSeconds int) error
}

var (
	// ErrDrainNodeTimeout indicates client is not able to drain the node before reaching timeout
	ErrDrainNodeTimeout = errors.New("drain node timeout")
)

func NewClient(ctx context.Context, cp config.IProvider, zl logger.IZapLogger) KubernetesClient {

	var kubeConfig *rest.Config

	switch cp.GetString(config.KubernetesRunMode) {
	case "InCluster":
		kubeConfig = getInclusterConfig()
	case "OutCluster":
		kubeConfig = getOutClusterConfig()
	default:
		zl.Info("No cluster mode selected. Configuring InCluster mode by default.")
		kubeConfig = getInclusterConfig()
	}

	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		panic(err.Error())
	}

	return KubernetesClient{logger: zl, Clientset: clientset, ctx: ctx}
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func getOutClusterConfig() *rest.Config {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	return config
}

func getInclusterConfig() *rest.Config {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	return config
}
