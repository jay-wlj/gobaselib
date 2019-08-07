package yf

import (
	"encoding/xml"
	"io"
	"strconv"
)

type StringMap map[string]string

type xmlMapEntry struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

// map本来已经是引用类型了，所以不需要 *Params
func (p StringMap) SetString(k, s string) StringMap {
	p[k] = s
	return p
}

func (p StringMap) GetString(k string) string {
	s, _ := p[k]
	return s
}

func (p StringMap) SetInt64(k string, i int64) StringMap {
	p[k] = strconv.FormatInt(i, 10)
	return p
}

func (p StringMap) SetInt(k string, i int) StringMap {
	p[k] = strconv.Itoa(i)
	return p
}

func (p StringMap) GetInt64(k string) int64 {
	i, _ := strconv.ParseInt(p.GetString(k), 10, 64)
	return i
}

// 判断key是否存在
func (p StringMap) ContainsKey(key string) bool {
	_, ok := p[key]
	return ok
}

func (m StringMap) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if len(m) == 0 {
		return nil
	}

	start.Name.Local = "xml" // 更改xml开始标签
	err := e.EncodeToken(start)
	if err != nil {
		return err
	}

	for k, v := range m {
		e.Encode(xmlMapEntry{XMLName: xml.Name{Local: k}, Value: v})
	}

	return e.EncodeToken(start.End())
}

func (m *StringMap) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	*m = StringMap{}
	for {
		var e xmlMapEntry

		err := d.Decode(&e)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		(*m)[e.XMLName.Local] = e.Value
	}
	return nil
}
