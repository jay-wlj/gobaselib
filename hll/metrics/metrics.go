package metrics

import (
	"fmt"
	"strconv"
	"time"

	"github.com/jay-wlj/gobaselib/hll/metrics/util"
	"github.com/prometheus/client_golang/prometheus"
)

func ReqHisgoram(uri string, statusCode int, err int, begin time.Time) {
	strRet := strconv.Itoa(err)
	if statusCode >= 500 {
		strRet = "1"
	}
	fmt.Println("appid=", appid)
	reqLatencyHisgoram.With(prometheus.Labels{
		"hll_appid":       appid,
		"hll_data_type":   "base",
		"hll_metric_type": "histogram",
		"route":           uri,
		"status":          fmt.Sprintf("%d", statusCode),
		"hll_env":         env,
		"hll_ip":          util.LocalIP(),
		"ret":             strRet,
	}).Observe(time.Now().Sub(begin).Seconds())
}

func DownRequestHisgorm(down_uri string, retErr string, begin time.Time) {

	downstreamRequestHisgorm.With(prometheus.Labels{
		"error":           retErr,
		"downstream_url":  down_uri,
		"hll_appid":       appid,
		"hll_data_type":   "business",
		"hll_metric_type": "histogram",
		"hll_env":         env,
		"hll_ip":          util.LocalIP(),
	}).Observe(time.Now().Sub(begin).Seconds())
}

func ApolloRequestHistgrom(down_uri string, retErr string, begin time.Time) {
	apolloRequestHistgrom.With(prometheus.Labels{
		"error":           retErr,
		"downstream_url":  down_uri,
		"hll_appid":       appid,
		"hll_data_type":   "business",
		"hll_metric_type": "histogram",
		"hll_env":         env,
		"hll_ip":          util.LocalIP(),
	}).Observe(time.Now().Sub(begin).Seconds())
}

func PanicCounter(uri string) {
	panicCounter.With(prometheus.Labels{
		"req_url":         uri,
		"hll_appid":       appid,
		"hll_data_type":   "business",
		"hll_metric_type": "counter",
		"hll_env":         env,
		"hll_ip":          util.LocalIP(),
	}).Inc()
}

func RedisHisgoram(cmd string, err int, begin time.Time) {
	redisLatencyHisgoram.With(prometheus.Labels{
		"cmd":             cmd,
		"error":           strconv.Itoa(err),
		"hll_appid":       appid,
		"resource":        "",
		"hll_data_type":   "base",
		"hll_metric_type": "histogram",
		"hll_env":         env,
		"hll_ip":          util.LocalIP(),
	}).Observe(time.Now().Sub(begin).Seconds())
}

func MysqlHisgoram(cmd, sql string, err int, begin time.Time) {
	mysqlLatencyHisgoram.With(prometheus.Labels{
		"cmd":             cmd,
		"sql":             sql,
		"error":           strconv.Itoa(err),
		"hll_appid":       appid,
		"resource":        "",
		"hll_data_type":   "base",
		"hll_metric_type": "histogram",
		"hll_env":         env,
		"hll_ip":          util.LocalIP(),
	}).Observe(time.Now().Sub(begin).Seconds())
}

func PromHisgoram(end_point string, begin time.Time) {
	promLatencyHisgoram.With(prometheus.Labels{
		"end_point":       end_point,
		"hll_appid":       appid,
		"hll_data_type":   "business",
		"hll_metric_type": "histogram",
		"hll_env":         env,
		"hll_ip":          util.LocalIP(),
	}).Observe(time.Now().Sub(begin).Seconds())
}

func ConsulRegCounter(reg_name, reg_id, errorStr string) {
	consulCounter.With(prometheus.Labels{
		"reg_name":        reg_name,
		"reg_id":          reg_id,
		"error":           errorStr,
		"hll_appid":       appid,
		"hll_data_type":   "business",
		"hll_metric_type": "histogram",
		"hll_env":         env,
		"hll_ip":          util.LocalIP(),
	}).Inc()
}
