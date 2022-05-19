package jpushclient

import (
	"encoding/json"
	"fmt"
	"github.com/jay-wlj/gobaselib/log"

	base "github.com/jay-wlj/gobaselib"
)

func NewAudience(user_ids []int64, ext_tags []string, ext_tags_and []string) (audience Audience, audience_old Audience) {
	str_user_ids := []string{}
	ball := true
	if len(user_ids) > 0 {
		for _, user_id := range user_ids {
			str_user_ids = append(str_user_ids, fmt.Sprintf("%v", user_id))
		}
		ball = false
		audience.SetAlias(str_user_ids) //用户ID推送应该使用别名
	}

	if len(ext_tags) > 0 {
		ball = false
		if len(user_ids) > 0 {
			//非广播消息,兼容已上线旧版本V1.7用户ID设置为tag的安卓
			old_tags := str_user_ids
			for _, tag := range ext_tags {
				old_tags = append(old_tags, tag)
			}
			audience_old.SetTag(old_tags)
		}
		audience.SetTag(ext_tags)
	} else if len(user_ids) > 0 {
		//非广播消息,兼容已上线旧版本V1.7用户ID设置为tag的安卓
		audience_old.SetTag(str_user_ids)
	}

	if len(ext_tags_and) > 0 {
		ball = false
		audience_old.SetTagAnd(ext_tags_and)
		audience.SetTagAnd(ext_tags_and)
	}

	if ball {
		audience.All()
	}
	return
}

func PushSendMsg(pusher *PushClient, bytes []byte) (msg_id string, err error) {
	retstr, err := pusher.Send(bytes)
	if err != nil {
		log.Errorf("pusher.Send failed!  err:%v", err)
		return
	}
	log.Infof("retstr:%v", retstr)
	map_ret := make(map[string]interface{})
	err = json.Unmarshal(base.Slice(retstr), &map_ret)
	if err != nil {
		log.Errorf("json.Unmarshal failed!  err:%v retstr:%v", err, retstr)
		return
	}
	msg_id, isexist := map_ret["msg_id"].(string)
	if !isexist {
		log.Errorf("map_ret[msg_id] not exists err:%v", retstr)
		err = fmt.Errorf("pushsendmsg failed! err:%v", retstr)
	}
	return
}

func PushSendSchedule(schedule *ScheduleClient, bytes []byte) (schedule_id string, err error) {
	retstr, err := schedule.Send(bytes)
	if err != nil {
		log.Errorf("pusher.Send failed!  err:%v", err)
		return
	}
	log.Infof("retstr:%v", retstr)
	map_ret := make(map[string]interface{})
	err = json.Unmarshal(base.Slice(retstr), &map_ret)
	if err != nil {
		log.Errorf("json.Unmarshal failed!  err:%v retstr:%v", err, retstr)
		return
	}
	schedule_id, isexist := map_ret["schedule_id"].(string)
	if !isexist {
		log.Errorf("map_ret[schedule_id] not exists,retstr:%v", retstr)
		err = fmt.Errorf("sendschedule failed! err:%v", retstr)
	}
	return
}

func JPushBind(secret string, appkey string, tag string, reg_id string) error {
	var reginfo DeviceRegiester
	reginfo.AddTag(tag)

	jclient := NewDevicesClient(secret, appkey)
	succ_ids, err := jclient.JDevicesRegister(&reginfo, []string{reg_id})
	if err != nil {
		log.Errorf("jclient.JDevicesRegister(%v,%v) failed! err:%v", reg_id, tag, err)
		return err
	}
	log.Errorf("regisinfo:%v succ_ids:%v", reginfo, succ_ids)
	if len(succ_ids) > 0 && succ_ids[0] == reg_id {
		return nil
	}
	return fmt.Errorf("regsiter failed, succ_id not contain %v", reg_id)
}

func JPushUnBind(secret string, appkey string, tag string, reg_id string) error {
	var reginfo DeviceRegiester
	reginfo.RemoveTag(tag)

	jclient := NewDevicesClient(secret, appkey)
	succ_ids, err := jclient.JDevicesRegister(&reginfo, []string{reg_id})
	if err != nil {
		log.Errorf("jclient.JDevicesRegister(%v) failed! err:%v", reginfo, err)
		return err
	}
	log.Errorf("regisinfo:%v succ_ids:%v", reginfo, succ_ids)
	if len(succ_ids) > 0 && succ_ids[0] == reg_id {
		return nil
	}
	return fmt.Errorf("regsiter failed, succ_id not contain %v", reg_id)
}

func JPushUserNotice(
	secret string,
	appkey string,
	user_ids []int64,
	ext_tags []string,
	ext_tags_and []string,
	content string,
	title string,
	extras map[string]interface{},
	pushDebug bool) (msg_id string, err error) {
	if 0 == len(user_ids) {
		return "", fmt.Errorf("usernotice not user_ids found")
	}
	pusher := NewPushClient(secret, appkey)

	audience, audience_old := NewAudience(user_ids, ext_tags, ext_tags_and)

	notice := NewNotice([]string{"android", "ios"}, title, content, extras)

	platform := NewPlatForm([]string{})

	payload := NewPushPayLoad(pushDebug)
	payload.SetAudience(&audience)
	payload.SetNotice(notice)
	payload.SetPlatform(platform)

	options := Option{}
	options.SetTimelive(3600 * 24 * 10)
	payload.SetOptions(&options)

	bytes, err := payload.ToBytes()
	if err != nil {
		log.Errorf("payload.ToBytes failed! err:%v", err)
		return
	}

	msg_id, err = PushSendMsg(pusher, bytes)
	if err != nil {
		return
	}

	if len(user_ids) > 0 {
		//非广播消息,兼容已上线旧版本V1.7用户ID设置为tag的安卓
		payload.SetAudience(&audience_old)
		bytes_old, err_old := payload.ToBytes()
		if err_old != nil {
			log.Errorf("old payload.ToBytes failed! err:%v", err_old)
			return
		}

		msg_id_old, err_old := PushSendMsg(pusher, bytes_old)
		log.Infof("send old msg msg_Id:%v err:%v", msg_id_old, err_old)
	}
	return
}

func JPushNotice(
	secret string,
	appkey string,
	platforms []string,
	alias []string,
	tags []string,
	tags_and []string,
	content string,
	title string,
	extras map[string]interface{},
	pushDebug bool) (msg_id string, err error) {

	pusher := NewPushClient(secret, appkey)
	audience := Audience{}
	ball := true
	if len(alias) > 0 {
		ball = false
		audience.SetAlias(alias)
	}
	if len(tags) > 0 {
		ball = false
		audience.SetTag(tags)
	}
	if len(tags_and) > 0 {
		ball = false
		audience.SetTagAnd(tags_and)
	}

	if ball {
		audience.All()
	}

	platform := NewPlatForm(platforms)
	if 0 == len(platforms) {
		platforms = append(platforms, "ios")
		platforms = append(platforms, "android")
	}

	notice := NewNotice(platforms, title, content, extras)

	payload := NewPushPayLoad(pushDebug)
	payload.SetAudience(&audience)
	payload.SetNotice(notice)
	payload.SetPlatform(platform)
	options := Option{}
	options.SetTimelive(3600 * 24 * 10)
	payload.SetOptions(&options)

	bytes, err := payload.ToBytes()
	if err != nil {
		log.Errorf("payload.ToBytes failed! err:%v", err)
		return
	}

	msg_id, err = PushSendMsg(pusher, bytes)

	return
}

func JPushScheduleNotice(
	secret string,
	appkey string,
	user_ids []int64,
	ext_tags []string,
	ext_tags_and []string,
	name string,
	enabled bool,
	start int64,
	end int64,
	time_ int,
	time_unit string,
	frequency int,
	points []string,
	platforms []string,
	content string,
	title string,
	extras map[string]interface{},
	pushDebug bool) (schedule_id string, err error) {

	schedule := NewScheduleClient(secret, appkey)

	audience, audience_old := NewAudience(user_ids, ext_tags, ext_tags_and)

	platform := NewPlatForm(platforms)
	if 0 == len(platforms) {
		platforms = append(platforms, "ios")
		platforms = append(platforms, "android")
	}

	notice := NewNotice(platforms, title, content, extras)

	payload := NewPushPayLoad(pushDebug)
	payload.SetAudience(&audience)
	payload.SetNotice(notice)
	payload.SetPlatform(platform)

	options := Option{}
	if end > start && (end-start) < 3600*24*10 {
		options.SetTimelive(int(end - start))
	} else {
		options.SetTimelive(3600 * 24 * 10)
	}
	if pushDebug {
		options.SetApns(false)
	} else {
		options.SetApns(true)
	}

	payload.SetOptions(&options)

	schedulepayload := NewSchedulePayLoad()
	schedulepayload.SetName(name)
	schedulepayload.SetEnabled(enabled)
	schedulepayload.SetPush(payload)

	trigger := NewTrigger(start, end, time_, time_unit, frequency, points)
	schedulepayload.SetTrigger(trigger)

	bytes, err := schedulepayload.ToBytes()
	if err != nil {
		log.Errorf("schedulepayload.ToBytes failed! err:%v", err)
		return
	}
	schedule_id, err = PushSendSchedule(schedule, bytes)
	if err != nil {
		log.Errorf("PushSendSchedule failed! err:%v", err)
	}

	if len(user_ids) > 0 {
		//非广播消息,兼容已上线旧版本V1.7用户ID设置为tag的安卓
		payload.SetAudience(&audience_old)
		schedulepayload.SetPush(payload)

		bytes_old, err_old := schedulepayload.ToBytes()
		if err_old != nil {
			log.Errorf("schedulepayload.ToBytes failed! err:%v", err_old)
			return
		}

		schedule_id_old, err_old := PushSendSchedule(schedule, bytes_old)
		if err_old != nil {
			log.Errorf("old PushSendSchedule failed! err:%v", err_old)
		} else {
			log.Infof("old schedepayload schedule_id:%v err:%v", schedule_id_old, err_old)
		}
	}

	return
}
