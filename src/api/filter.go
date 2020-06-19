package api

import (
	"strconv"
	"strings"
	"time"

	"github.com/emicklei/go-restful"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/xdhuxc/kubernetes-transform/src/config"
	"github.com/xdhuxc/kubernetes-transform/src/pkg"
)

func (bc *BaseController) page(req *restful.Request, resp *restful.Response,
	chain *restful.FilterChain) {

	if req.Request.Method != "GET" {
		chain.ProcessFilter(req, resp)
		return
	}

	var pageSize int64 = 10
	var cpage int64 = 1
	if ps, err := strconv.ParseInt(req.QueryParameter("limit"), 10, 64); err == nil &&
		ps > 0 {
		pageSize = ps
	}
	if p, err := strconv.ParseInt(req.QueryParameter("page"), 10, 64); err == nil &&
		p > 0 {
		cpage = p
	}

	offset := (cpage - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	page := pkg.Page{
		PageSize: pageSize,
		Offset:   offset,
		Page:     cpage,
		Query:    req.QueryParameter("query"),
	}

	switch req.QueryParameter("sort") {
	case "asc":
		page.Sort = "asc"
	default:
		page.Sort = "desc"
	}

	switch req.QueryParameter("order_by") {
	case "name":
		page.OrderBy = "name"
	default:
		page.OrderBy = "update_time"
	}

	req.SetAttribute("page", page)

	chain.ProcessFilter(req, resp)
}

func (bc *BaseController) metrics(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	start := time.Now()
	duration := float64(time.Since(start)) / float64(time.Second)

	httpRequestTotal.With(prometheus.Labels{
		"method":   req.Request.Method,
		"endpoint": req.Request.URL.Path,
		"code":     strconv.Itoa(resp.StatusCode()),
		"env":      config.GetConfig().Env,
	}).Inc()

	httpRequestDuration.With(prometheus.Labels{
		"method":   req.Request.Method,
		"endpoint": req.Request.URL.Path,
		"code":     strconv.Itoa(resp.StatusCode()),
		"env":      config.GetConfig().Env,
	}).Observe(duration)

	chain.ProcessFilter(req, resp)
}

func (bc *BaseController) check(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	if strings.Contains(req.Request.URL.Path, pkg.RequestTypeTransform) {
		err := bc.bs.TransformService.Check()
		pkg.WriteResponse(resp, pkg.TransformServiceCheckError, err)
		return
	}

	if strings.Contains(req.Request.URL.Path, pkg.RequestTypeRestore) {
		err := bc.bs.RestoreService.Check()
		pkg.WriteResponse(resp, pkg.RestoreServiceCheckError, err)
		return
	}

	if strings.Contains(req.Request.URL.Path, pkg.RequestTypeSave) {
		err := bc.bs.SaveService.Check()
		pkg.WriteResponse(resp, pkg.SaveServiceCheckError, err)
		return
	}

	chain.ProcessFilter(req, resp)
}
