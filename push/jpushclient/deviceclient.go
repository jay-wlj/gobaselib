package jpushclient

import (
	//"encoding/base64"
	//"errors"
	//"strings"
	"encoding/json"
	"fmt"
	"github.com/jay-wlj/gobaselib/log"
)

const (
	//SUCCESS_FLAG  = "msg_id"
	HOST_NAME_DEVICES_SSL = "https://device.jpush.cn/v3/devices/"
	//BASE64_TABLE          = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
)

//var base64Coder = base64.NewEncoding(BASE64_TABLE)

type DevicesClient struct {
	MasterSecret string
	AppKey       string
	AuthCode     string
	BaseUrl      string
}

type Deviceresponse struct {
	Tags   []string `json:"tags"`
	Alias  string   `json:"alias"`
	Mobile int64    `json:"mobile"`
}

type TagsAddAndRemove struct {
	TagsAdd    interface{} `json:"add,omitempty"`
	TagsRemove interface{} `json:"remove,omitempty"`
}

type DeviceRegiester struct {
	Tags   interface{} `json:"tags,omitempty"`
	Alias  interface{} `json:"alias,omitempty"`
	Mobile interface{} `json:"mobile,omitempty"`
}

func (this *DeviceRegiester) AddTag(tag string) {
	var tagsandremove TagsAddAndRemove
	if nil != this.Tags {
		tagsandremove = this.Tags.(TagsAddAndRemove)
	}

	var tagsadd []string
	if nil != tagsandremove.TagsAdd {
		tagsadd = tagsandremove.TagsAdd.([]string)
	}
	tagsadd = append(tagsadd, tag)
	tagsandremove.TagsAdd = tagsadd
	this.Tags = tagsandremove
}

func (this *DeviceRegiester) RemoveTag(tag string) {
	var tagsandremove TagsAddAndRemove
	if nil != this.Tags {
		tagsandremove = this.Tags.(TagsAddAndRemove)
	}

	var tagsremove []string
	if nil != tagsandremove.TagsRemove {
		tagsremove = tagsandremove.TagsRemove.([]string)
	}
	tagsremove = append(tagsremove, tag)
	tagsandremove.TagsRemove = tagsremove
	this.Tags = tagsandremove
}

func (this *DeviceRegiester) SetAlias(alias string) {
	this.Alias = alias
}

func (this *DeviceRegiester) SetMobile(mobile string) {
	this.Mobile = mobile
}

func (this *DeviceRegiester) ToBytes(mobile string) {

}

func NewDevicesClient(secret, appkey string) *DevicesClient {
	auth := "Basic " + base64Coder.EncodeToString([]byte(appkey+":"+secret))
	devicesclient := &DevicesClient{secret, appkey, auth, HOST_NAME_DEVICES_SSL}
	return devicesclient
}

func (this *DevicesClient) Query(registration_id string) (ret string, err error) {
	ret, err = SendGet(this.BaseUrl+registration_id, make(map[string]string), this.AuthCode)

	if err != nil {
		log.Errorf("SendGet ret:%v err:%v", ret, err)
		return
	}
	log.Infof("query(%v) ret:%v err:%v", registration_id, ret, err)
	map_ret := make(map[string]interface{})
	err = json.Unmarshal([]byte(ret), &map_ret)
	code, isexist := map_ret["error"]
	if isexist {
		return "", fmt.Errorf("error not found, code:%v ret:%v", code, ret)
	}

	return
}

func (this *DevicesClient) JDevicesRegister(regisinfo *DeviceRegiester, registration_ids []string) (succ_ids []string, err error) {
	if 0 == len(registration_ids) {
		return
	}
	data, err := json.Marshal(regisinfo)
	if err != nil {
		log.Errorf("json.marshal(%v) failed! err:%v", regisinfo, err)
		return
	}

	for _, registration_id := range registration_ids {
		ret, err := SendPostBytes2(this.BaseUrl+registration_id, data, this.AuthCode)
		if err != nil {
			log.Errorf("1.registertags:%v registration_id:%v failed! ret:%v err:%v", data, registration_id, ret, err)
		} else {
			succ_ids = append(succ_ids, registration_id)
			log.Errorf("2.registertags:%v registration_id:%v failed! ret:%v err:%v", data, registration_id, ret, err)
		}
	}

	return
}
