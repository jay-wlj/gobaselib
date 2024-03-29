package yf

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	base "github.com/jay-wlj/gobaselib"
)

func TokenCheck(token string) (user_id, user_type, expire_time int64, err error) {
	return token_check_server(TokenConfig.AccountServer, token, TokenConfig.AccountTimeout)
}

// TODO: 检查Token过期时间。
func token_check_server(account_server, token string, timeout time.Duration) (user_id, user_type, expire_time int64, err error) {
	uri := account_server + "/man/account/token/check"
	headers := make(map[string]string)
	headers["Host"] = "account.yunbay.com"
	headers["X-Not-Use-Proxy"] = "true"
	headers["X-YF-Token"] = token

	res := base.HttpGetJson(uri, headers, timeout)
	//log.Infof("request [%s] status: %d", res.ReqDebug, res.StatusCode)

	if res.StatusCode != 200 {
		log.Debugf("request [%s] failed! err: %v", res.ReqDebug, res.Error)
		err = res.Error
		if res.Error == nil {
			err = fmt.Errorf("ERR_SERVER_ERROR")
		}
		return
	}

	if !res.Ok {
		log.Errorf("request [%s] failed! reason: %s", res.ReqDebug, res.Reason)
		err = fmt.Errorf(res.Reason)
		return
	}

	jn_user_id := res.Data["user_id"].(json.Number)
	jn_user_type := res.Data["user_type"].(json.Number)
	jn_expire_time := res.Data["expire_time"].(json.Number)
	user_id, err = jn_user_id.Int64()

	log.Infof("----------------tockcheck(%v, %v, %v)---------------err:%v", jn_user_id, user_id, jn_expire_time, err)
	if err != nil {
		return
	}
	expire_time, err = jn_expire_time.Int64()
	if err != nil {
		return
	}
	user_type, err = jn_user_type.Int64()
	if err != nil {
		return
	}
	return
}

type tokenSet struct {
	Token string `json:"token"`
}

type tokenRes struct {
	Reason string   `json:"reason"`
	Ok     bool     `json:"ok"`
	Data   tokenSet `json:"data"`
}

func TokenGet(account_server string, user_id int, timeout time.Duration) (token string, err error) {
	//uri := account_server + "/account/man/token?user_id=" + strconv.Itoa(user_id)
	uri := account_server + "/account/man/get_token?user_id=" + strconv.Itoa(user_id)
	headers := make(map[string]string)
	//	headers["Host"] = "account.lapianapp.com"
	headers["Host"] = "account.nicefilm.com"
	headers["X-Not-Use-Proxy"] = "true"

	res := base.HttpGet(uri, headers, timeout)
	log.Infof("request [%s] status: %d", res.ReqDebug, res.StatusCode)

	if res.StatusCode != 200 {
		log.Errorf("request [%s] failed!,code:%d err: %v", res.ReqDebug, res.StatusCode, res.Error)
		err = fmt.Errorf("ERR_SERVER_ERROR")
		return
	}
	set := new(tokenRes)
	err = json.Unmarshal(res.RawBody, set)
	if err != nil {
		log.Errorf("unmarsh [%v],error:%s", string(res.RawBody), err.Error())
		return
	}
	token = set.Data.Token
	if token == "" {
		err = fmt.Errorf("ERR_SERVER_ERROR")
	}

	return
}
