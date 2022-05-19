package yf

import "gobaselib/log"

const (
	ERR_ARGS_INVALID         string = "ERR_ARGS_INVALID"         // 参数错误
	ERR_SERVER_ERROR         string = "ERR_SERVER_ERROR"         // 服务器错误
	ERR_HASH_INVALID         string = "ERR_HASH_INVALID"         // hash不正确
	ERR_CONTENT_TYPE_INVALID string = "ERR_CONTENT_TYPE_INVALID" // content-type非法
	ERR_FILE_NOT_SUPPORT     string = "ERR_FILE_NOT_SUPPORT"     // 文件不支持
	ERR_SIGN_ERROR           string = "ERR_SIGN_ERROR"           // 签名错误
	ERR_TOKEN_INVALID        string = "ERR_TOKEN_INVALID"        // token非法
	ERR_IMGCODE_ERROR        string = "ERR_IMGCODE_ERROR"        // 图片验证码错误
	ERR_OPEN_INPUT_FILE      string = "ERR_OPEN_INPUT_FILE"      // 打开文件错误
	ERR_OBJECT_NOT_FOUND     string = "ERR_OBJECT_NOT_FOUND"     // 对象没找到
	ERR_HTTP_FORBIDDEN       string = "ERR_HTTP_FORBIDDEN"       // http forbidden
	ERR_ARGS_MISSING         string = "ERR_ARGS_MISSING"         // 参数缺失
	ERR_NOT_FOUND            string = "ERR_NOT_FOUND"            // 没有找到
	ERR_NOT_EXISTS           string = "ERR_NOT_EXISTS"           // 不存在的记录
	DATA_NOT_MOTIFIED        string = "DATA_NOT_MOTIFIED"        // 数据不能修改
	DATA_NOT_EXIST           string = "DATA_NOT_EXIST"           // 数据不存在
	DATA_NOT_SUPPORT         string = "DATA_NOT_SUPPORT"         // 不支持
	ERR_EMAIL_INVALID        string = "ERR_EMAIL_INVALID"        // 非法的邮箱地址
	ERR_ACTION_UNSUPPORTED   string = "ERR_ACTION_UNSUPPORTED"
	ERR_OBJECT_RECOMMENDED   string = "ERR_OBJECT_RECOMMENDED"
)

func GetStatusCode(reason string) int {
	switch reason {
	case ERR_CONTENT_TYPE_INVALID, ERR_ARGS_INVALID, ERR_HASH_INVALID, ERR_FILE_NOT_SUPPORT:
		return 400
	case ERR_SIGN_ERROR, ERR_TOKEN_INVALID:
		return 401
	case ERR_SERVER_ERROR:
		return 500
	default:
		log.Errorf("Unknow Reason: %s", reason)
		return 500
	}
}

func IsArgsMissing(err error) bool {
	if err.Error() == ERR_ARGS_MISSING {
		return true
	}
	return false
}

func IsErrorNotFound(err error) bool {
	if err.Error() == "not found" {
		return true
	}
	return false
}
