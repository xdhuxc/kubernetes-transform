package api

import (
	"github.com/xdhuxc/kubernetes-transform/src/model"
	"net/http"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"github.com/xdhuxc/kubernetes-transform/src/pkg"
)

type ServiceController struct {
	*BaseController
}

func newServiceController(bc *BaseController) *ServiceController {
	tags := []string{"kubernetes-transform-service"}
	sc := &ServiceController{bc}

	sc.ws.Route(sc.ws.POST("/save").
		To(sc.Save).
		Doc("save resources of target cluster").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "OK", pkg.Result{}).
		Returns(http.StatusBadRequest, "ERROR", pkg.Result{}))

	sc.ws.Route(sc.ws.POST("/restore").
		To(sc.Resotre).
		Doc("restore resources to target cluster").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "OK", pkg.Result{}).
		Returns(http.StatusBadRequest, "ERROR", pkg.Result{}))

	sc.ws.Route(sc.ws.POST("/transform").
		To(sc.Transform).
		Doc("transform resources from source cluster to target cluster").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "OK", pkg.Result{}).
		Returns(http.StatusBadRequest, "ERROR", pkg.Result{}))

	return sc
}

func (sc *ServiceController) Save(req *restful.Request, resp *restful.Response) {

	sc.bs.SaveService.Save()

}

func (sc *ServiceController) Transform(req *restful.Request, resp *restful.Response) {
	var tr []model.TransformRequest

	err := req.ReadEntity(&tr)
	if err != nil {
		pkg.WriteResponse(resp, pkg.TransformRequestError, err)
		return
	}

	sc.bs.TransformService.Transform(tr)

}

func (sc *ServiceController) Resotre(req *restful.Request, resp *restful.Response) {

	sc.bs.RestoreService.Restore()

}
