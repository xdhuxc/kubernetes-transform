package api

import (
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsController struct {
	*BaseController
}

var httpRequestDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "The HTTP request latencies in seconds",
		Buckets: nil,
	}, []string{"method", "endpoint", "code", "env"},
)

var httpRequestTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_request_total",
		Help: "The total number of http request to a route",
	}, []string{"method", "endpoint", "code", "env"},
)

func init() {
	// 注册监控指标
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(httpRequestTotal)
}

func newMetricsController(baseController *BaseController) *MetricsController {
	return &MetricsController{baseController}
}

func (mc *MetricsController) ServeHTTP(httpwriter http.ResponseWriter, httpRequest *http.Request) {
	// 添加各种指标

	promhttp.Handler().ServeHTTP(httpwriter, httpRequest)
}

func metrics(c *restful.Container, bc *BaseController) {
	mc := newMetricsController(bc)

	c.Handle("/metrics", mc)
}
