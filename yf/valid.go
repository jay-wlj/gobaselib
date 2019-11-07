package yf

import (
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	"gopkg.in/go-playground/validator.v9"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func Var(v interface{}, tag string) error {
	return validate.Var(v, tag)
}

func Valid(req interface{}) error {

	// 只校验结构体对象参数
	if err := validate.Struct(req); err != nil {
		switch err.(type) {
		case *validator.InvalidValidationError: // 参数非struct类型 不判断
			err = nil
		case validator.ValidationErrors:
			glog.Error("Valid fail! err=", (err.(validator.ValidationErrors)).Error())
			return err
		}
	}

	return nil
}

func ValidTel(telNum string) bool {
	regular := "^((13[0-9])|(14[5,7])|(15[0-3,5-9])|(17[0,3,5-8])|(18[0-9])|166|198|199|(147))\\d{8}$"

	reg := regexp.MustCompile(regular)
	return reg.MatchString(telNum)
}

func UnmarshalReq(c *gin.Context, req interface{}) bool {
	var err error
	switch c.Request.Method {
	case "GET":
		err = c.ShouldBindQuery(req)
	default:
		err = c.ShouldBindJSON(req)
	}

	if err != nil {
		//if err := base.CheckQueryJsonField(c, &req); err != nil {
		glog.Info("UnmarshalReq args invalid! err=", err)
		JSON_FailEx(c, ERR_ARGS_INVALID, err.Error())
		return false
	}

	return true
}
