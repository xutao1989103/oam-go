package apiserver

import (
	"bytes"
	"context"
	"fmt"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	oamapis "github.com/crossplane/oam-kubernetes-runtime/apis/core"
	oamv1alpha2 "github.com/crossplane/oam-kubernetes-runtime/pkg/controller/v1alpha2"
	"github.com/emicklei/go-restful"
	"github.com/xutao1989103/oam-go/api/v1alpha1"
	"github.com/xutao1989103/oam-go/controllers"
	configv1alpha2 "github.com/xutao1989103/oam-go/pkg/apis/config/v1alpha2"
	"github.com/xutao1989103/oam-go/pkg/apis/k8s/api"
	"github.com/xutao1989103/oam-go/pkg/apiserver/config"
	"github.com/xutao1989103/oam-go/pkg/client/k8s"
	"github.com/xutao1989103/oam-go/pkg/informers"
	"k8s.io/apimachinery/pkg/runtime/schema"
	urlruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog"
	"k8s.io/klog/klogr"
	"net"
	"net/http"
	rt "runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"strings"
	"time"
)

type Server struct {
	APIServer       *http.Server
	Config          *config.Config
	container       *restful.Container
	Clients         k8s.Clients
	InformerFactory informers.InformerFactory
}

func (s *Server) PrepareRun() error {

	s.container = restful.NewContainer()
	s.container.Filter(logRequestAndResponse)
	s.container.Router(restful.CurlyRouter{})
	s.container.RecoverHandler(func(panicReason interface{}, httpWriter http.ResponseWriter) {
		logStackOnRecover(panicReason, httpWriter)
	})

	s.installOAMAPIs()

	for _, ws := range s.container.RegisteredWebServices() {
		klog.V(2).Infof("%s", ws.RootPath())
	}

	s.APIServer.Handler = s.container

	return nil
}

func (s *Server) OAMControllerRun(stopCh <-chan struct{}) (err error) {
	logf.SetLogger(zap.Logger(false))
	klog.V(0).Info("setting up manager")
	mgr, err := manager.New(s.Clients.Config(), manager.Options{Port: 8443})
	if err != nil {
		klog.Fatalf("unable to set up overall controller manager: %v", err)
	}

	oamapis.AddToScheme(mgr.GetScheme())
	oamv1alpha2.Setup(mgr, logging.NewLogrLogger(ctrl.Log.WithName("oam-controller")))

	v1alpha1.AddToScheme(mgr.GetScheme())
	pipelineReconciler := &controllers.PipelineReconciler{
		Client: mgr.GetClient(),
		Log: klogr.New().WithName("Pipeline-log"),
		Scheme: mgr.GetScheme(),
	}
	pipelineReconciler.SetupWithManager(mgr);

	return mgr.Start(stopCh)
}

func (s *Server) Run(stopCh <-chan struct{}) (err error) {

	//err = s.waitForResourceSync(stopCh)
	//if err != nil {
	//	return err
	//}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-stopCh
		_ = s.APIServer.Shutdown(ctx)
	}()

	klog.V(0).Infof("Start listening on %s", s.APIServer.Addr)
	if s.APIServer.TLSConfig != nil {
		err = s.APIServer.ListenAndServeTLS("", "")
	} else {
		err = s.APIServer.ListenAndServe()
	}

	return err
}

func (s *Server) installOAMAPIs() {
	urlruntime.Must(configv1alpha2.AddToContainer(s.container, s.Config))
	urlruntime.Must(api.AddToContainer(s.container, s.Config, s.Clients))

}

func (s *Server) waitForResourceSync(stopCh <-chan struct{}) error {
	klog.V(0).Info("Start cache objects")

	discoveryClient := s.Clients.Kubernetes().Discovery()
	_, apiResourcesList, err := discoveryClient.ServerGroupsAndResources()
	if err != nil {
		return err
	}

	isResourceExists := func(resource schema.GroupVersionResource) bool {
		for _, apiResource := range apiResourcesList {
			if apiResource.GroupVersion == resource.GroupVersion().String() {
				for _, rsc := range apiResource.APIResources {
					if rsc.Name == resource.Resource {
						return true
					}
				}
			}
		}
		return false
	}

	// resources we have to create informer first
	k8sGVRs := []schema.GroupVersionResource{
		{Group: "", Version: "v1", Resource: "namespaces"},
		{Group: "", Version: "v1", Resource: "nodes"},
		{Group: "", Version: "v1", Resource: "resourcequotas"},
		{Group: "", Version: "v1", Resource: "pods"},
		{Group: "", Version: "v1", Resource: "services"},
		{Group: "", Version: "v1", Resource: "persistentvolumeclaims"},
		{Group: "", Version: "v1", Resource: "secrets"},
		{Group: "", Version: "v1", Resource: "configmaps"},

		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles"},
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings"},
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"},
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"},

		{Group: "apps", Version: "v1", Resource: "deployments"},
		{Group: "apps", Version: "v1", Resource: "daemonsets"},
		{Group: "apps", Version: "v1", Resource: "replicasets"},
		{Group: "apps", Version: "v1", Resource: "statefulsets"},
		{Group: "apps", Version: "v1", Resource: "controllerrevisions"},

		{Group: "storage.k8s.io", Version: "v1", Resource: "storageclasses"},

		{Group: "batch", Version: "v1", Resource: "jobs"},
		{Group: "batch", Version: "v1beta1", Resource: "cronjobs"},

		{Group: "extensions", Version: "v1beta1", Resource: "ingresses"},

		{Group: "autoscaling", Version: "v2beta2", Resource: "horizontalpodautoscalers"},

		{Group: "networking.k8s.io", Version: "v1", Resource: "networkpolicies"},
	}

	for _, gvr := range k8sGVRs {
		if !isResourceExists(gvr) {
			klog.Warningf("resource %s not exists in the cluster", gvr)
		} else {
			_, err := s.InformerFactory.KubernetesSharedInformerFactory().ForResource(gvr)
			if err != nil {
				klog.Errorf("cannot create informer for %s", gvr)
				return err
			}
		}
	}

	s.InformerFactory.KubernetesSharedInformerFactory().Start(stopCh)
	s.InformerFactory.KubernetesSharedInformerFactory().WaitForCacheSync(stopCh)

	klog.V(0).Info("Finished caching objects")

	return nil

}

func logStackOnRecover(panicReason interface{}, w http.ResponseWriter) {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("recover from panic situation: - %v\r\n", panicReason))
	for i := 2; ; i += 1 {
		_, file, line, ok := rt.Caller(i)
		if !ok {
			break
		}
		buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
	}
	klog.Errorln(buffer.String())

	headers := http.Header{}
	if ct := w.Header().Get("Content-Type"); len(ct) > 0 {
		headers.Set("Accept", ct)
	}

	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Internal server error"))
}

func logRequestAndResponse(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	start := time.Now()
	chain.ProcessFilter(req, resp)

	// Always log error response
	logWithVerbose := klog.V(4)
	if resp.StatusCode() > http.StatusBadRequest {
		logWithVerbose = klog.V(0)
	}

	logWithVerbose.Infof("%s - \"%s %s %s\" %d %d %dms",
		getRequestIP(req),
		req.Request.Method,
		req.Request.URL,
		req.Request.Proto,
		resp.StatusCode(),
		resp.ContentLength(),
		time.Since(start)/time.Millisecond,
	)
}

func getRequestIP(req *restful.Request) string {
	address := strings.Trim(req.Request.Header.Get("X-Real-Ip"), " ")
	if address != "" {
		return address
	}

	address = strings.Trim(req.Request.Header.Get("X-Forwarded-For"), " ")
	if address != "" {
		return address
	}

	address, _, err := net.SplitHostPort(req.Request.RemoteAddr)
	if err != nil {
		return req.Request.RemoteAddr
	}

	return address
}
