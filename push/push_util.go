package push

import (
	"encoding/json"
	"fmt"
	"github.com/jay-wlj/gobaselib/push/jpushclient"
	"strings"
	"time"
	"unicode"

	"github.com/jie123108/glog"
	"github.com/zwczou/jpush"
)

type PushConfig struct {
	MapAppKeys   map[string]string
	MongoUrl     string
	TimeOut      time.Duration
	PushDebug    bool
	PushDebugTag string
}

var pushconfig *PushConfig

func InitPush(map_appkeys map[string]string, mongourl string, timeout time.Duration, pushDebug bool, pushdebugtag string) (err error) {
	pushconfig = &PushConfig{map_appkeys, mongourl, timeout, pushDebug, pushdebugtag}
	return nil
}

func QueryUserAppKey(user_ids []int64) (map[string][]int64, error) {
	dao, err := NewJPushBindDao()
	if nil != err {
		glog.Errorf("NewJPushBindDao(%v) failed! err:%v", user_ids, err)
		return make(map[string][]int64), err
	}

	valueinfos, err := dao.FindByUserIds(user_ids)
	if nil != err {
		glog.Errorf("QueryUserAppKey.dao.Find failed, err:%v", err)
		return make(map[string][]int64), err
	}

	glog.Infof("find user_ids:%v keys:%v", user_ids, valueinfos)
	map_keys := make(map[string][]int64)
	map_keys_once := make(map[string]map[int64]bool)
	map_user_find_keys := make(map[int64]bool)

	for _, valueinfo := range valueinfos {
		appkey := valueinfo.AppKey
		user_id := valueinfo.UserId
		map_user_find_keys[user_id] = true
		if nil == map_keys[appkey] {
			map_keys[appkey] = []int64{}
			map_keys_once[appkey] = make(map[int64]bool)
		}
		if !map_keys_once[appkey][user_id] {
			map_keys[appkey] = append(map_keys[appkey], user_id)
			map_keys_once[appkey][user_id] = true
		}
	}

	//找不到属于那一个应用的用户,所有平台广播一下
	for _, user_id := range user_ids {
		if !map_user_find_keys[user_id] {
			glog.Infof("user_id:%v not find appkey, broadcat all appkey", user_id)
			for appkey, _ := range pushconfig.MapAppKeys {
				if nil == map_keys[appkey] {
					map_keys[appkey] = []int64{}
				}
				map_keys[appkey] = append(map_keys[appkey], user_id)
			}
		}
	}
	return map_keys, err
}

func PushBind(appkey string, user_id int64, reg_id string) error {
	if nil == pushconfig || nil == pushconfig.MapAppKeys {
		return fmt.Errorf("Push model not inited or not set appkeys!")
	}
	str_user_id := fmt.Sprintf("%v", user_id)
	/*secret, isexist := pushconfig.MapAppKeys[appkey]
	if !isexist {
		glog.Errorf("JPushBind Failed! AppKey not found")
		return fmt.Errorf("JPushBind Failed! AppKey not found")
	}*/

	//改为客户端注册用户ID,服务器只保存注册信息
	/*err := jpushclient.JPushBind(secret, appkey, str_user_id, reg_id)
	if nil != err {
		return err
	}*/

	dao, err := NewJPushBindDao()
	if nil != err {
		glog.Errorf("NewJPushBindDao(%v,%v,%v) failed! err:%v", reg_id, str_user_id, appkey, err)
		return err
	}

	err = dao.Upsert(reg_id, appkey, user_id)
	if nil != err {
		glog.Errorf("dao.Upsert(%v,%v,%v) failed! err:%v", reg_id, str_user_id, appkey, err)
		return err
	}

	return err
}

func PushUnBind(appkey string, user_id int64, reg_id string) error {
	if nil == pushconfig || nil == pushconfig.MapAppKeys {
		return fmt.Errorf("Push model not inited or not set appkeys!")
	}
	str_user_id := fmt.Sprintf("%v", user_id)

	/*secret, isexist := pushconfig.MapAppKeys[appkey]
	if !isexist {
		glog.Errorf("JPushBind Failed! AppKey not found")
		return fmt.Errorf("JPushBind Failed! AppKey not found")
	}*/

	//改为客户端注册用户ID,服务器只保存注册信息
	/*
		err := jpushclient.JPushUnBind(secret, appkey, str_user_id, reg_id)
		if nil != err {
			glog.Errorf("jpushun(%v,%v)bind failed! err:%v", reg_id, str_user_id, err)
			return err
		}*/

	dao, err := NewJPushBindDao()
	if nil != err {
		glog.Errorf("NewJPushBindDao(%v,%v,%v) failed! err:%v", reg_id, str_user_id, appkey, err)
		return err
	}

	err = dao.DeleteByRegId(reg_id)
	if nil != err {
		glog.Errorf("dao.Delete(%v,%v,%v) failed! err:%v", reg_id, str_user_id, appkey, err)
		return err
	}

	return err
}

func Query(reg_id string) (ret int, err error) {
	for appkey, secret := range pushconfig.MapAppKeys {
		devicesclient := jpushclient.NewDevicesClient(secret, appkey)
		ret, err := devicesclient.Query(reg_id)
		glog.Errorf("ret:%v err:%v", ret, err)
	}
	return
}

func UserNotice(
	user_ids []int64,
	ext_tags []string,
	ext_tags_and []string,
	content string,
	title string,
	extras map[string]interface{}) (msg_id string, err error) {
	if 0 == len(user_ids) {
		return "", fmt.Errorf("usernotice not user_ids found")
	}

	if nil == pushconfig || nil == pushconfig.MapAppKeys {
		return "", fmt.Errorf("Push model not inited or not set appkeys!")
	}

	map_keys, err := QueryUserAppKey(user_ids)
	if nil != err {
		return "", err
	}

	if pushconfig.PushDebug {
		ext_tags_and = append(ext_tags_and, pushconfig.PushDebugTag)
	}
	for appkey, uids := range map_keys {
		//appkey := "66eaf7bd2b39b8ecdcd966a6"
		secret, isexist := pushconfig.MapAppKeys[appkey]
		if !isexist {
			glog.Errorf("JPushBind Failed! AppKey(%v) uids(%v) not found", appkey, uids)
			continue
			//return "", fmt.Errorf("JPushBind Failed! AppKey not found")
		}

		msg_id, err = jpushclient.JPushUserNotice(secret, appkey, uids, ext_tags, ext_tags_and, content, title, extras, pushconfig.PushDebug)
		if nil != err {
			glog.Errorf("jpushclient.jpushusernotice(%v,%v,%v,%v)", uids, content, title, extras)
			//return "", err
		}

	}

	return
}

func BroadNotice(
	platforms []string,
	alias []string,
	tags []string,
	tags_and []string,
	content string,
	title string,
	extras map[string]interface{}) (msg_id string, err error) {
	if nil == pushconfig || nil == pushconfig.MapAppKeys {
		return "", fmt.Errorf("Push model not inited or not set appkeys!")
	}

	if pushconfig.PushDebug {
		tags_and = append(tags_and, pushconfig.PushDebugTag)
	}

	for appkey, secret := range pushconfig.MapAppKeys {
		msg_id, err = jpushclient.JPushNotice(secret, appkey, platforms, alias, tags, tags_and, content, title, extras, pushconfig.PushDebug)
		if nil != err {
			glog.Errorf("jpushclient.BroadNotice(%v,%v) failed!", secret, appkey)
		}
	}

	return
}

/*
func PlatfromTagsAndTagsAnd1(
	platform string,
	mp_platform_tags map[string][]string,
	mp_platform_tags_and map[string][]string) ([]string, []string) {
	all := "all"
	tags, is_tags := mp_platform_tags[platform]
	tags_and, is_tags_and := mp_platform_tags_and[platform]
	all_tags, is_all_tags := mp_platform_tags[all]
	all_tags_and, is_all_tags_and := mp_platform_tags_and[all]
	if is_tags || is_tags_and {
		if is_tags && !is_tags_and {
			if is_all_tags_and {
				tags_and = all_tags_and
			} else {
				tags_and = []string{}
			}
		} else if is_tags_and && !is_tags {
			if is_all_tags {
				tags = all_tags
			} else {
				tags = []string{}
			}
		}
	} else {
		if is_all_tags {
			tags = all_tags
		} else {
			tags = []string{}
		}

		if is_all_tags_and {
			tags_and = all_tags_and
		} else {
			tags_and = []string{}
		}
	}
	return tags, tags_and
}*/

// 优化代码
// 如果tagMap, tagAndMap存在platform的key直接返回
// 否在返回all的key
func PlatfromTagsAndTagsAnd(platform string, tagMap, tagAndMap map[string][]string) (tags []string, tagAnds []string) {
	defaultKey := "all"
	tags = tagMap[platform]
	tagAnds = tagAndMap[platform]
	if tags == nil {
		tags = tagMap[defaultKey]
	}
	if tagAnds == nil {
		tagAnds = tagMap[defaultKey]
	}
	return
}

func ScheduleNoticeV2(
	user_ids []int64,
	mp_platform_tags map[string][]string,
	mp_platform_tags_and map[string][]string,
	name string,
	enabled bool,
	start int64,
	end int64,
	time_ int,
	time_unit string,
	frequency int,
	points []string,
	content string,
	title string,
	extras map[string]interface{}) (schedule_id_android string, schedule_id_ios string, err error) {
	// glog.Infof("xxxx --------- 0001 ------------")
	// 因为ios有bug, 所以分开推送.
	// if 0 == len(mp_platform_tags) && 0 == len(mp_platform_tags_and) {
	// 	//glog.Infof("mp_platform_tags: %v, mp_platform_tags_and: %v", mp_platform_tags, mp_platform_tags_and)
	// 	ext_tags := []string{}
	// 	ext_tags_and := []string{}
	// 	sid, err_ := ScheduleNotice(user_ids, ext_tags, ext_tags_and, name, enabled, start, end, time_, time_unit, frequency, points, []string{}, content, title, extras)
	// 	schedule_id_android = sid
	// 	schedule_id_ios = sid
	// 	err = err_
	// 	return
	// }

	//glog.Infof("xxxx --------- 0002 ------------")

	android := "android"
	ios := "ios"
	android_tags, android_tags_and := PlatfromTagsAndTagsAnd(android, mp_platform_tags, mp_platform_tags_and)
	ios_tags, ios_tags_and := PlatfromTagsAndTagsAnd(ios, mp_platform_tags, mp_platform_tags_and)
	schedule_id_android, err = ScheduleNotice(user_ids, android_tags, android_tags_and, name, enabled, start, end, time_, time_unit, frequency, points, []string{android}, content, title, extras)
	if err != nil {
		glog.Errorf("schedulenotice android(%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v)  failed! err:%v",
			user_ids, android_tags, android_tags_and, name, enabled, start, end, time_, time_unit,
			frequency, points, content, title, extras, err)
		return
	}

	//glog.Infof("xxxx --------- 0003 ------------")
	push := extras["push"]
	// ios老版本, push节点只能解析string类型的信息(老版本的信息).所以为了兼容, 把消息改回原来的格式了.
	body, _ := json.Marshal(push)
	extras["push"] = string(body)
	// glog.Infof("xxxx push msg to ios: %v", extras)
	schedule_id_ios, err = ScheduleNotice(user_ids, ios_tags, ios_tags_and, name, enabled, start, end, time_, time_unit, frequency, points, []string{ios}, content, title, extras)
	return schedule_id_android, schedule_id_ios, err
}

// 将特殊逻辑从jpush库里面提炼出来
func NewNotification(platform *jpush.Platform, title, content string, extras map[string]interface{}) *jpush.Notification {
	notice := &jpush.Notification{
		Alert: content,
	}
	if platform.Has(jpush.Ios) {
		notice.Ios = &jpush.IosNotification{
			Alert: map[string]string{
				"title": title,
				"body":  content,
			},
			//Badge:  1,
			Badge:  "+1",
			Extras: extras,
		}
	}
	if platform.Has(jpush.Android) {
		notice.Android = &jpush.AndroidNotification{
			Title:  title,
			Alert:  content,
			Extras: extras,
		}
	}
	// winphone 已经被微软放弃
	return notice
}

// 发送定时消息
// 结构体直接赋值，没必要采用setAttribute
func JPushScheduleNotice(secret, appkey string,
	user_ids []int64, ext_tags, ext_tags_and []string,
	name string, enabled bool,
	start int64, end int64, time_ int, time_unit string, frequency int, points []string,
	platforms []string,
	content string, title string, extras map[string]interface{}, pushDebug bool) (schedule_id string, err error) {

	platform := jpush.NewPlatform().All()
	if len(platforms) > 0 {
		platform.Add(platforms...)
	}

	// 最长时间为10天
	timeLive := end - start
	if (end < start) || (end-start) > 864000 {
		timeLive = 864000
	}

	// 将[]int转成[]string
	userIdStrs := strings.FieldsFunc(fmt.Sprint(user_ids), func(r rune) bool {
		return !unicode.IsNumber(r)
	})
	audience := jpush.NewAudience().SetAlias(userIdStrs...).SetTag(ext_tags...).SetTagAnd(ext_tags_and...)

	payload := &jpush.Payload{
		Platform:     platform,
		Audience:     audience,
		Notification: NewNotification(platform, title, content, extras),
		Options: &jpush.Options{
			ApnsProduction: !pushDebug,
			TimeLive:       int(timeLive),
		},
	}

	body, _ := json.Marshal(payload)
	glog.Infof("push content: %s", body)

	schedulePayload := &jpush.SchedulePayload{
		Name:    name,
		Enabled: enabled,
		Push:    payload,
		Trigger: &jpush.Trigger{
			Single: &jpush.Single{
				Time: time.Unix(start, 0).Format("2006-01-02 15:04:05"),
			},
		},
	}
	client := jpush.NewJpushClient(appkey, secret)
	schedule_id, err = client.ScheduleCreate(schedulePayload)
	glog.Infof("jpush schedule %s - %s", schedule_id, err)
	//client.ScheduleDelete(schedule_id)
	return
}

func ScheduleNotice(
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
	extras map[string]interface{}) (schedule_id string, err error) {
	if nil == pushconfig || nil == pushconfig.MapAppKeys {
		return "", fmt.Errorf("Push model not inited or not set appkeys!")
	}

	if pushconfig.PushDebug {
		ext_tags_and = append(ext_tags_and, pushconfig.PushDebugTag)
	}
	if len(user_ids) > 0 {
		map_keys, err := QueryUserAppKey(user_ids)
		if nil != err {
			return "", err
		}

		for appkey, uids := range map_keys {
			//appkey := "66eaf7bd2b39b8ecdcd966a6"
			secret, isexist := pushconfig.MapAppKeys[appkey]
			if !isexist {
				glog.Errorf("JPushBind Failed! AppKey(%v) uids(%v) not found", appkey, uids)
				continue
				//return 0, fmt.Errorf("JPushBind Failed! AppKey not found")
			}
			glog.Infof("JPushScheduleNotice(%v, %v, %v, %v, %v, %v, %v, %v, %v,%v, %v, %v, %v, %v, %v, %v, %v, %v)",
				secret, appkey, uids, ext_tags, ext_tags_and, name, enabled, start, end,
				time_, time_unit, frequency, points, platforms, content, title, extras, pushconfig.PushDebug)

			schedule_id, err = JPushScheduleNotice(secret, appkey, uids, ext_tags, ext_tags_and, name, enabled, start, end,
				time_, time_unit, frequency, points, platforms, content, title, extras, pushconfig.PushDebug)
			if nil != err {
				glog.Errorf("jpushclient.JPushScheduleNotice(%v,%v,%v,%v)", uids, content, title, extras)
				//return 0, err
			}
		}

	} else {
		for appkey, secret := range pushconfig.MapAppKeys {
			glog.Infof("JPushScheduleNotice(%v, %v, %v, %v, %v, %v, %v, %v, %v,%v, %v, %v, %v, %v, %v, %v, %v, %v)",
				secret, appkey, user_ids, ext_tags, ext_tags_and, name, enabled, start, end,
				time_, time_unit, frequency, points, platforms, content, title, extras, pushconfig.PushDebug)

			schedule_id, err = JPushScheduleNotice(secret, appkey, user_ids, ext_tags, ext_tags_and, name, enabled, start, end,
				time_, time_unit, frequency, points, platforms, content, title, extras, pushconfig.PushDebug)
			if nil != err {
				glog.Errorf("jpushclient.BroadNotice(%v,%v) failed!", secret, appkey)
			}
		}

	}

	return schedule_id, err
}
