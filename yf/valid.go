package yf

import (	
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
		case *validator.InvalidValidationError:	// 参数非struct类型 不判断			
			err = nil 							
		case validator.ValidationErrors:
			glog.Error("Valid fail! err=", (err.(validator.ValidationErrors)).Error())	
			return err
		}				
	}		

	return nil
}