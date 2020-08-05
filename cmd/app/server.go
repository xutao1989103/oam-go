package app

import (
	kconfig "github.com/kiali/kiali/config"
	"github.com/spf13/cobra"
	options2 "github.com/xutao1989103/oam-go/cmd/app/options"
	apiserverconfig "github.com/xutao1989103/oam-go/pkg/apiserver/config"
	"github.com/xutao1989103/oam-go/utils/signals"
)

func NewOAMServerCommand() *cobra.Command {

	serverRunOptions := options2.NewServerRunOptions()

	conf, tryError := apiserverconfig.TryLoadFromDisk()
	if tryError == nil {
		serverRunOptions.Config = conf
	}

	cmd := &cobra.Command{
		Use:  "oam-server",
		Long: "this is an oam server implement by golang",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := Run(serverRunOptions, signals.SetupSignalHandler())
			if err != nil {
				return err
			}
			return nil
		},
	}

	return cmd
}

func Run(options *options2.ServerRunOptions, stopCh <-chan struct{}) error {
	initialize(options)

	server, err := options.NewServer(stopCh)
	if err != nil {
		return err
	}

	go server.OAMControllerRun(stopCh)

	err = server.PrepareRun()
	if err != nil {
		return err
	}

	return server.Run(stopCh)

	return nil
}

func initialize(options *options2.ServerRunOptions) {
	config := kconfig.NewConfig()

	config.API.Namespaces.Exclude = []string{"istio-system", "kubesphere*", "kube*"}
	config.InCluster = true

}
