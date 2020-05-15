package api

import (
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/emicklei/go-restful"
	"k8s.io/klog/v2"

	"github.com/xdhuxc/kubernetes-transform/src/client"
	"github.com/xdhuxc/kubernetes-transform/src/config"
)

type Router struct {
	container *restful.Container
	bs        *BaseController
}

func NewRouter() *Router {
	mysqldb, err := client.NewMySQLClient(config.GetConfig().Database)
	if err != nil {
		fmt.Printf("new mysql client error: %v\n", err)
		return nil
	}

	baseController := NewBaseController(mysqldb)
	container := restful.NewContainer()
	container.Add(baseController.ws)

	metrics(container, baseController)
	swagger(container, config.GetConfig().Address)
	staticWs(container)

	baseController.ws.Filter(baseController.metrics)
	baseController.ws.Filter(baseController.page)

	r := &Router{
		container: container,
		bs:        baseController,
	}

	return r
}

func (r *Router) Run() error {
	fmt.Printf("start http server at : %s", config.GetConfig().Address)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", config.GetConfig().Address),
		Handler:      r.container,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	return server.ListenAndServe()
}

func staticWs(c *restful.Container) {
	ws := new(restful.WebService)
	ws.Route(ws.GET("/static/{subpath:*}").To(staticFromPathParam))
	c.Add(ws)
}

func staticFromPathParam(req *restful.Request, resp *restful.Response) {
	actual := path.Join("./static", req.PathParameter("subpath"))
	klog.Errorf("serving %s ... (from %s)\n", actual, req.PathParameter("subpath"))
	http.ServeFile(
		resp.ResponseWriter,
		req.Request,
		actual)
}
