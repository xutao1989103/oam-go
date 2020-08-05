package api

import (
	"context"
	"github.com/emicklei/go-restful"
	"github.com/xutao1989103/oam-go/pkg/apiserver/config"
	"github.com/xutao1989103/oam-go/pkg/client/k8s"
)

func AddToContainer(c *restful.Container, config *config.Config, client k8s.Clients) error {
	webService := restful.WebService{}
	webService.Path("/proxy").Produces(restful.MIME_JSON)

	webService.Route(webService.GET("").
		Doc("Proxy to K8s api").
		To(func(request *restful.Request, response *restful.Response) {
			result := client.Kubernetes().AppsV1().RESTClient().
				Verb(request.Request.Method).
				RequestURI(request.QueryParameter("URL")).
				Do(context.TODO())
			obj, _ := result.Get()
			response.WriteAsJson(obj)
		}))

	c.Add(&webService)
	return nil
}
