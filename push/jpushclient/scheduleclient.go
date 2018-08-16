package jpushclient

import (
	//"encoding/base64"
	"errors"
	"strings"
)

const (
	HOST_NAME_SCHEDULE_SSL = "https://api.jpush.cn/v3/schedules"
)

type ScheduleClient struct {
	MasterSecret string
	AppKey       string
	AuthCode     string
	BaseUrl      string
}

func NewScheduleClient(secret, appkey string) *ScheduleClient {
	//base64
	auth := "Basic " + base64Coder.EncodeToString([]byte(appkey+":"+secret))
	pusher := &ScheduleClient{secret, appkey, auth, HOST_NAME_SCHEDULE_SSL}
	return pusher
}

func (this *ScheduleClient) Send(data []byte) (string, error) {
	return this.SendPushBytes(data)
}

func (this *ScheduleClient) SendPushString(content string) (string, error) {
	ret, err := SendPostString(this.BaseUrl, content, this.AuthCode)
	if err != nil {
		return ret, err
	}
	if strings.Contains(ret, "schedule_id") {
		return ret, nil
	} else {
		return "", errors.New(ret)
	}
}

func (this *ScheduleClient) SendPushBytes(content []byte) (string, error) {
	//ret, err := SendPostBytes(this.BaseUrl, content, this.AuthCode)
	ret, err := SendPostBytes2(this.BaseUrl, content, this.AuthCode)
	if err != nil {
		return ret, err
	}
	if strings.Contains(ret, "schedule_id") {
		return ret, nil
	} else {
		return "", errors.New(ret)
	}
}
