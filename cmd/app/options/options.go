package options

import (
	"fmt"
	"github.com/xutao1989103/oam-go/pkg/apiserver"
	"github.com/xutao1989103/oam-go/pkg/apiserver/config"
	"github.com/xutao1989103/oam-go/pkg/client/k8s"
	"net/http"
)

type ServerRunOptions struct {
	ConfigFile              string
	GenericServerRunOptions *GenericServerRunOptions
	*config.Config
	DebugMode bool
}

func NewServerRunOptions() *ServerRunOptions {
	s := &ServerRunOptions{
		ConfigFile:              "/etc/oam",
		Config:                  config.NewConfig(),
		GenericServerRunOptions: NewGenericServerRunOptions(),
		DebugMode:               false,
	}

	return s
}

func (options *ServerRunOptions) NewServer(stopCh <-chan struct{}) (*apiserver.Server, error) {

	server := apiserver.Server{
		Config: options.Config,
	}

	allClient, err := k8s.NewKubernetesClient(options.K8sOptions)
	if err != nil {
		return nil, err
	}
	server.Clients = allClient

	apiServer := &http.Server{
		Addr: fmt.Sprintf(":%d", options.GenericServerRunOptions.InsecurePort),
	}

	server.APIServer = apiServer

	return &server, nil
}

type GenericServerRunOptions struct {
	// server bind address
	BindAddress string

	// insecure port number
	InsecurePort int

	// secure port number
	SecurePort int

	// tls cert file
	TlsCertFile string

	// tls private key file
	TlsPrivateKey string
}

func NewGenericServerRunOptions() *GenericServerRunOptions {
	// create default server run confoptions
	s := GenericServerRunOptions{
		BindAddress:   "0.0.0.0",
		InsecurePort:  9090,
		SecurePort:    0,
		TlsCertFile:   "",
		TlsPrivateKey: "",
	}

	return &s
}
