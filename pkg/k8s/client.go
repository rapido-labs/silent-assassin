package k8s

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

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
	logger logger.IZapLogger
}

type IKubernetesClient interface {
	GetNodes(labelSelector string) (*v1.NodeList, error)
	GetNode(name string) (v1.Node, error)
	GetPodsInNode(name string) ([]v1.Pod, error)
	DeletePod(name, namespace string) error
	DeleteNode(name string) error
	UpdateNode(node v1.Node) error
}

func NewClient(cp config.IProvider, zl logger.IZapLogger) KubernetesClient {

	var kubeConfig *rest.Config

	switch cp.GetString(config.KubernetesRunMode) {
	case "InCluster":
		kubeConfig = getInclusterConfig()
	case "OutCluster":
		kubeConfig = getOutClusterConfig()
	default:
		panic(fmt.Sprintf("No cluster mode specified in %v", config.KubernetesRunMode))
	}

	clientset, err := kubernetes.NewForConfig(kubeConfig)

	if err != nil {
		panic(err.Error())
	}

	return KubernetesClient{logger: zl, Clientset: clientset}
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
