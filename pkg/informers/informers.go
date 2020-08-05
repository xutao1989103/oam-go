package informers

import (
	istioclient "istio.io/client-go/pkg/clientset/versioned"
	istioinformers "istio.io/client-go/pkg/informers/externalversions"
	k8sinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"time"
)

const defaultResync = 600 * time.Second

type InformerFactory interface {
	Start(stopCh <-chan struct{})
	KubernetesSharedInformerFactory() k8sinformers.SharedInformerFactory
	IstioSharedInformerFactory() istioinformers.SharedInformerFactory
}

type informerFactories struct {
	informerFactory      k8sinformers.SharedInformerFactory
	istioInformerFactory istioinformers.SharedInformerFactory
}

func NewInformerFactories(client kubernetes.Interface, istioClient istioclient.Interface) InformerFactory {
	factories := &informerFactories{}
	if client != nil {
		factories.informerFactory = k8sinformers.NewSharedInformerFactory(client, defaultResync)
	}

	if istioClient != nil {
		factories.istioInformerFactory = istioinformers.NewSharedInformerFactory(istioClient, defaultResync)
	}
	return factories
}

func (f *informerFactories) KubernetesSharedInformerFactory() k8sinformers.SharedInformerFactory {
	return f.informerFactory
}

func (f *informerFactories) IstioSharedInformerFactory() istioinformers.SharedInformerFactory {
	return f.istioInformerFactory
}

func (f *informerFactories) Start(stopCh <-chan struct{}) {
	if f.informerFactory != nil {
		f.informerFactory.Start(stopCh)
	}
	if f.istioInformerFactory != nil {
		f.istioInformerFactory.Start(stopCh)
	}
}
