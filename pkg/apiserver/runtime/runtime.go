package runtime

import (
	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	ApiRootPath = "/kapis"
)

var Container = restful.NewContainer()

type ContainerBuilder []func(c *restful.Container) error

const MimeMergePatchJson = "application/merge-patch+json"
const MimeJsonPatchJson = "application/json-patch+json"

func init() {
	restful.RegisterEntityAccessor(MimeMergePatchJson, restful.NewEntityAccessorJSON(restful.MIME_JSON))
	restful.RegisterEntityAccessor(MimeJsonPatchJson, restful.NewEntityAccessorJSON(restful.MIME_JSON))
}

func NewWebService(gv schema.GroupVersion) *restful.WebService {
	webService := restful.WebService{}
	webService.Path(ApiRootPath + "/" + gv.String()).Produces(restful.MIME_JSON)
	return &webService
}

func (cb *ContainerBuilder) AddToContainer(c *restful.Container) error {
	for _, f := range *cb {
		if err := f(c); err != nil {
			return err
		}
	}
	return nil
}

func (cb *ContainerBuilder) Register(funcs ...func(*restful.Container) error) {
	for _, f := range funcs {
		*cb = append(*cb, f)
	}
}
