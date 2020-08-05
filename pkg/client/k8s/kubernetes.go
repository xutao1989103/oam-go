package k8s

import (
	istioclient "istio.io/client-go/pkg/clientset/versioned"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Clients interface {
	Master() string
	Kubernetes() kubernetes.Interface
	Istio() istioclient.Interface
	Config() *rest.Config
}

type kubernetesClient struct {
	k8s    kubernetes.Interface
	istio  istioclient.Interface
	master string
	config *rest.Config
}

func NewKubernetesClient(options *K8sOptions) (Clients, error) {
	config, err := clientcmd.BuildConfigFromFlags("", options.KubeConfig)
	if err != nil {
		return nil, err
	}
	config.QPS = options.QPS
	config.Burst = 10
	config.GroupVersion = &schema.GroupVersion{}

	var kClient kubernetesClient

	kClient.config = config
	kClient.k8s, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &kClient, nil
}

func (k *kubernetesClient) Kubernetes() kubernetes.Interface {
	return k.k8s
}

func (k *kubernetesClient) Master() string {
	return k.master
}

func (k *kubernetesClient) Istio() istioclient.Interface {
	return k.istio
}

func (k *kubernetesClient) Config() *rest.Config {
	return k.config
}
