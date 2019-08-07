package yf

import (
	"encoding/json"
	"fmt"

	"github.com/jie123108/glog"
	nsq "github.com/nsqio/go-nsq"
)

type NsqMsg interface {
	Topic() string // 返回topic
}
type Nsq struct {
	mps map[string]*nsq.Producer
}

func NewNsq(mqurls []string) (r *Nsq) {
	mps := make(map[string]*nsq.Producer)
	for _, v := range mqurls {
		mps[v] = initProducer(v)
	}
	r = &Nsq{mps}
	glog.Error("mqurls:", r.mps, " mqurls:", mqurls)
	return
}

// 发布消息
func (m *Nsq) PublishMsg(v NsqMsg) error {
	content, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return m.asyncPublishMsg(v.Topic(), content)
}

// 异步发布消息
func (m *Nsq) AsyncPublishMsg(v NsqMsg) error {
	content, err := json.Marshal(v)
	if err != nil {
		return err
	}
	go m.asyncPublishMsg(v.Topic(), content)
	return err
}

func (m *Nsq) asyncPublishMsg(topic string, body []byte) (err error) {
	for k, v := range m.mps {
		err = publish(v, topic, string(body))
		if err != nil {
			glog.Errorf("publih fail! err=%v, url:%v", err, k)
			continue
		}
		return
	}
	glog.Error("all nsqd publih fail! topic:", topic, " body:", string(body))

	return
}

// 初始化生产者
func initProducer(str string) (producer *nsq.Producer) {
	var err error
	glog.Info("address: ", str)
	producer, err = nsq.NewProducer(str, nsq.NewConfig())
	if err != nil {
		glog.Error("initProducer fail! err=", err)
		panic(err)
	}
	return
}

//发布消息
func publish(v *nsq.Producer, topic string, message string) error {
	var err error
	if v != nil {
		if message == "" { //不能发布空串，否则会导致error
			return nil
		}
		err = v.Publish(topic, []byte(message)) // 发布消息
		return err
	}
	return fmt.Errorf("producer is nil err=%v", err)
}
