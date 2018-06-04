package jpushclient

import (
//"encoding/base64"
//"errors"
//"strings"
)

const (
	//SUCCESS_FLAG  = "msg_id"
	HOST_NAME_REPORT_SSL = "https://report.jpush.cn/v3/received"
	//BASE64_TABLE  = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
)

//var base64Coder = base64.NewEncoding(BASE64_TABLE)

type ReportClient struct {
	MasterSecret string
	AppKey       string
	AuthCode     string
	BaseUrl      string
}

func NewReportClient(secret, appKey string) *ReportClient {
	//base64
	auth := "Basic " + base64Coder.EncodeToString([]byte(appKey+":"+secret))
	report := &ReportClient{secret, appKey, auth, HOST_NAME_REPORT_SSL}
	return report
}

func (this *ReportClient) Query(msg_ids string) {
	return
}
