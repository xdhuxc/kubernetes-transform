package api

import (
	"net/http"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"

	"github.com/xdhuxc/kubernetes-transform/src/pkg"
)

type healthController struct {
	*BaseController
}

func newHealthCheckController(bc *BaseController) *healthController {
	tags := []string{"health"}
	hcc := &healthController{bc}

	hcc.ws.Route(hcc.ws.GET("/health").
		To(hcc.Get).
		Doc("health check").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "OK", pkg.Result{}).
		Returns(http.StatusBadRequest, "ERROR", pkg.Result{}))

	return hcc
}

func (hcc *healthController) Get(req *restful.Request, resp *restful.Response) {
	result, err := hcc.bs.HealthService.Get()
	if err != nil {
		pkg.WriteResponse(resp, pkg.HealthError, err)
		return
	}

	_ = resp.WriteEntity(pkg.NewResult(0, nil, result))
}
