package mt

import "github.com/jie123108/glog"

const (
	ERR_ARGS_INVALID         string = "ERR_ARGS_INVALID"
	ERR_SERVER_ERROR         string = "ERR_SERVER_ERROR"
	ERR_HASH_INVALID         string = "ERR_HASH_INVALID"
	ERR_CONTENT_TYPE_INVALID string = "ERR_CONTENT_TYPE_INVALID"
	ERR_FILE_NOT_SUPPORT     string = "ERR_FILE_NOT_SUPPORT"
	ERR_SIGN_ERROR           string = "ERR_SIGN_ERROR"
	ERR_TOKEN_INVALID        string = "ERR_TOKEN_INVALID"
	ERR_OPEN_INPUT_FILE      string = "ERR_OPEN_INPUT_FILE"
	ERR_OBJECT_NOT_FOUND     string = "ERR_OBJECT_NOT_FOUND"
	ERR_HTTP_FORBIDDEN       string = "ERR_HTTP_FORBIDDEN"
	ERR_ARGS_MISSING         string = "ERR_ARGS_MISSING"
	DATA_NOT_MOTIFIED        string = "DATA_NOT_MOTIFIED"
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
		glog.Errorf("Unknow Reason: %s", reason)
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
