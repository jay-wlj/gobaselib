package base

import (
	"fmt"
	"github.com/jay-wlj/gobaselib/log"
	"strconv"
	"strings"
	"time"
)

type ABuYunStatus struct {
	CurrentIP  string
	UsedSecond int
	RestSecond int
}

func GetABuYunStatus() (status *ABuYunStatus) {
	status = &ABuYunStatus{}
	status_uri := "http://proxy.abuyun.com/current-ip"
	res := HttpGet(status_uri, nil, time.Second*10)
	if res.StatusCode != 200 {
		log.Errorf("request [%s] failed! status: %d, err: %v", res.ReqDebug, res.StatusCode, res.Error)
		return
	}
	body := strings.TrimSpace(string(res.RawBody))
	arr := strings.SplitN(body, ",", 3)
	if len(arr) != 3 {
		return
	}

	status.CurrentIP = strings.TrimSpace(arr[0])
	status.UsedSecond, _ = strconv.Atoi(strings.TrimSpace(arr[1]))
	status.RestSecond, _ = strconv.Atoi(strings.TrimSpace(arr[2]))
	return
}

// Unix time
var pre_fetch_time int64
var g_status *ABuYunStatus

/**
 * 获取http://www.abuyun.com/代理的状态，带缓存。5s获取一次。
 */
func GetABuYunStatusEx() *ABuYunStatus {
	now := time.Now().Unix()
	diff := int(now - pre_fetch_time)
	// log.Infof(" now [%d] - pre_fetch_time [%d] = %d", now, pre_fetch_time, diff)

	if g_status == nil || diff >= 5 || g_status.RestSecond < diff {
		pre_fetch_time = now
		// log.Infof("------------------ Get New Status ------------------")
		g_status = GetABuYunStatus()
		return g_status
	}
	// log.Infof("------------------ Get Status From cache ------------------")
	status := &ABuYunStatus{}
	status.CurrentIP = g_status.CurrentIP
	status.UsedSecond = g_status.UsedSecond + diff
	status.RestSecond = g_status.RestSecond - diff

	return status
}

func PrintABuYunStatus() {
	status := GetABuYunStatusEx()
	fmt.Printf("Current IP: %s, Used Second: %ds, Can Use Second: %ds\n",
		status.CurrentIP, status.UsedSecond, status.RestSecond)
}
