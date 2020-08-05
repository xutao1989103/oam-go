package v1alpha2

import (
	"github.com/emicklei/go-restful"
	"github.com/xutao1989103/oam-go/pkg/apiserver/config"
	"github.com/xutao1989103/oam-go/pkg/apiserver/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	GroupName = "config.oam.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

func AddToContainer(c *restful.Container, config *config.Config) error {
	webservice := runtime.NewWebService(GroupVersion)

	webservice.Route(webservice.GET("/confoptions/configz").
		Doc("Information about the server configuration").
		To(func(request *restful.Request, response *restful.Response) {
			response.WriteAsJson(config.ToMap())
		}))

	c.Add(webservice)
	return nil
}
