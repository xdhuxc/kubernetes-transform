package api

import (
	"strconv"
	"strings"

	"github.com/emicklei/go-restful"
	"github.com/jinzhu/gorm"

	"github.com/xdhuxc/kubernetes-transform/src/config"
	"github.com/xdhuxc/kubernetes-transform/src/service"
)

type BaseController struct {
	db     *gorm.DB
	bs     *service.BaseService
	ws     *restful.WebService
	config config.Config
}

func NewBaseController(db *gorm.DB) (*BaseController, error) {
	ws := new(restful.WebService)
	ws.Path("/kubernetes/api/v1").
		Consumes(restful.MIME_XML, restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	c := config.GetConfig()
	bs, err := service.NewBaseService(c, db)
	if err != nil {
		return nil, err
	}

	baseController := &BaseController{
		db:     db,
		bs:     bs,
		ws:     ws,
		config: c,
	}

	newHealthCheckController(baseController)
	newServiceController(baseController)

	return baseController, nil
}

func (bc *BaseController) extract(req *restful.Request) (string, []int, string, error) {
	if config.GetConfig().Debug {
		return "debuger", []int{0}, "admin", nil
	}

	user := req.HeaderParameter("x-xdhuxc-user")
	role := req.HeaderParameter("x-xdhuxc-role")

	var gids []int
	for _, v := range strings.Split(strings.Trim(req.HeaderParameter("x-xdhuxc-group"), "[]"), " ") {
		gid, err := strconv.Atoi(v)
		if err != nil {
			return "", nil, "", err
		}
		gids = append(gids, gid)
	}

	return user, gids, role, nil
}
