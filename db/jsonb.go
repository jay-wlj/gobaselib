package db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

type Jsonb struct {
	json.RawMessage
}

func (j Jsonb) Value() (driver.Value, error) {
	if len(j.RawMessage) == 0 {
		return nil, nil
	}
	return j.MarshalJSON()
}

func (j *Jsonb) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}
	return json.Unmarshal(bytes, j)
}

type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	valueString, err := json.Marshal(j)
	return string(valueString), err
}

func (j *JSONB) Scan(value interface{}) error {
	src, ok := value.([]byte)
	if !ok {
		return errors.New("Type assertion .([]byte) failed!")
	}

	// 查看是否需要转义
	if len(src) > 0 && src[0] == '"' {
		s, err := strconv.Unquote(string(src))
		if err != nil {
			return err
		}
		//fmt.Println("s=",s)
		src = []byte(s)
	}

	if err := json.Unmarshal(src, &j); err != nil {
		return err
	}
	//fmt.Println("json Scan end src=", src)
	return nil
}

type JsonbArray []interface{}

func (j JsonbArray) Value() (driver.Value, error) {
	valueString, err := json.Marshal(j)
	return string(valueString), err
}

func (j *JsonbArray) Scan(value interface{}) error {
	src, ok := value.([]byte)
	if !ok {
		return errors.New("Type assertion .([]byte) failed!")
	}

	if err := json.Unmarshal(src, &j); err != nil {
		return err
	}
	//fmt.Println("json Scan end src=", src)
	return nil
}
