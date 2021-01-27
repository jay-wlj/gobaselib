package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	appid string
	env   string
)

func Init(appId, _env string) {
	appid = appId
	env = _env
	return
}

var (
	PrometheusHandler = func(c *gin.Context) {
		promhttp.Handler().ServeHTTP(c.Writer, c.Request)
	}

	reqLatencyHisgoram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "hllci_api_request_seconds",
		Help: "the total number of processed events",
	}, []string{"route", "status", "ret", "hll_data_type", "hll_metric_type", "hll_appid", "hll_env", "hll_ip"})

	redisLatencyHisgoram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "hllci_redis_request_seconds",
		Help: "the total number of redis processed events",
	}, []string{"cmd", "error", "hll_data_type", "hll_metric_type", "hll_appid", "hll_env", "hll_ip", "resource"})

	mysqlLatencyHisgoram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "hllci_mysql_request_seconds",
		Help: "the total number of redis processed events",
	}, []string{"cmd", "sql", "resource", "error", "hll_data_type", "hll_metric_type", "hll_appid", "hll_env", "hll_ip"})

	promLatencyHisgoram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "hllci_prom_request_seconds",
		Help: "the total number of redis processed events",
	}, []string{"end_point", "hll_data_type", "hll_metric_type", "hll_appid", "hll_env", "hll_ip"})

	panicCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "hllci_runtime_panic_count",
		Help: "the total number of canceled parallelism job",
	}, []string{"req_url", "hll_data_type", "hll_metric_type", "hll_appid", "hll_env", "hll_ip"})

	downstreamRequestHisgorm = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "hllci_downstream_http_request_seconds",
		Help: "the total number of redis processed events",
	}, []string{"hll_data_type", "error", "downstream_url", "hll_metric_type", "hll_appid", "hll_env", "hll_ip"})

	consulCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "hllci_consul_reg_count",
		Help: "the total number of reg in the consul",
	}, []string{"reg_name", "reg_id", "error", "hll_data_type", "hll_metric_type", "hll_appid", "hll_env", "hll_ip"})

	consulDiscoverCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "hllci_consul_discover_count",
		Help: "the total number of discover in the consul",
	}, []string{"service", "error", "err_msg", "hll_data_type", "hll_metric_type", "hll_appid", "hll_env", "hll_ip"})

	apolloRequestHistgrom = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "hllci_apollo_http_request_seconds",
		Help: "the total number of redis processed events",
	}, []string{"hll_data_type", "error", "downstream_url", "hll_metric_type", "hll_appid", "hll_env", "hll_ip"})
)
