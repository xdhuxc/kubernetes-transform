package api

import (
	"net/http"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"

	"github.com/xdhuxc/kubernetes-transform/src/pkg"
)

type healthCheckController struct {
	*BaseController
}

func newHealthCheckController(bc *BaseController) *healthCheckController {
	tags := []string{"hi"}
	hcc := &healthCheckController{bc}

	hcc.ws.Route(hcc.ws.GET("/health").
		To(hcc.Get).
		Doc("health check").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "OK", pkg.Result{}).
		Returns(http.StatusBadRequest, "ERROR", pkg.Result{}))

	return hcc
}

func (hcc *healthCheckController) Get(req *restful.Request, resp *restful.Response) {
	result, err := hcc.bs.HealthCheckService.Get()
	if err != nil {
		pkg.WriteResponse(resp, pkg.HiError, err)
		return
	}

	_ = resp.WriteEntity(pkg.NewResult(0, nil, result))
}
