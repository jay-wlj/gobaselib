package jpushclient

import (
	"encoding/json"
	"fmt"
	"time"
)

type SchedulePayLoad struct {
	Name    string      `json:"name"`
	Enabled bool        `json:"enabled"`
	Trigger interface{} `json:"trigger,omitempty"`
	Push    interface{} `json:"push,omitempty"`
}

type Trigger struct {
	Single_     interface{} `json:"single,omitempty"`
	Periodical_ interface{} `json:"periodical,omitempty"`
}

type Single struct {
	Time string `json:"time"`
}

type Periodical struct {
	Start     string   `json:"start"`
	End       string   `json:"end"`
	Time      string   `json:"time"`
	TimeUnit  string   `json:"time_unit"`
	Frequency int      `json:"frequency"`
	Point     []string `json:"point"`
}

const (
	TIME_UNIT_DAY   = "day"
	TIME_UNIT_WEEK  = "week"
	TIME_UNIT_MONTH = "month"
)

const (
	WEEK_MONDAY    = "MON"
	WEEK_TUESDAY   = "TUE"
	WEEK_WEDNESDAY = "WED"
	WEEK_THURSDAY  = "THU"
	WEEK_FRIDAY    = "FRI"
	WEEK_SATURDAY  = "SAT"
	WEEK_SUNDAY    = "SUN"
)

const (
	DAY_1  = "01"
	DAY_2  = "02"
	DAY_3  = "03"
	DAY_4  = "04"
	DAY_5  = "05"
	DAY_6  = "06"
	DAY_7  = "07"
	DAY_8  = "08"
	DAY_9  = "09"
	DAY_10 = "10"
	DAY_11 = "11"
	DAY_12 = "12"
	DAY_13 = "13"
	DAY_14 = "14"
	DAY_15 = "15"
	DAY_16 = "16"
	DAY_17 = "17"
	DAY_18 = "18"
	DAY_19 = "19"
	DAY_20 = "20"
	DAY_21 = "21"
	DAY_22 = "22"
	DAY_23 = "23"
	DAY_24 = "24"
	DAY_25 = "25"
	DAY_26 = "26"
	DAY_27 = "27"
	DAY_28 = "28"
	DAY_29 = "29"
	DAY_30 = "30"
	DAY_31 = "31"
)

func NewSchedulePayLoad() *SchedulePayLoad {
	return &SchedulePayLoad{Enabled: true}
}

func NewTrigger(
	start int64,
	end int64,
	time_ int,
	time_unit string,
	frequency int,
	points []string) (trigger *Trigger) {
	trigger = &Trigger{}
	/*if start > 0 && end > 0 {
		periodical := Periodical{}
		periodical.SetStartTime(start)
		periodical.SetEndTime(end)
		periodical.SetTime(time_)
		periodical.SetTimeUnit(time_unit)
		periodical.SetFrequency(frequency)
		for _, p := range points {
			periodical.AddPoint(p)
		}
		trigger.SetPeriodical(&periodical)
	} else {*/
	single := Single{}
	single.SetTime(start)
	trigger.SetSingle(&single)
	//}
	return trigger
}

func (this *SchedulePayLoad) SetName(name string) {
	this.Name = name
}

func (this *SchedulePayLoad) SetEnabled(enabled bool) {
	this.Enabled = enabled
}

func (this *SchedulePayLoad) SetTrigger(trigger *Trigger) {
	this.Trigger = trigger
}

func (this *SchedulePayLoad) SetPush(push *PayLoad) {
	this.Push = push
}

func (this *SchedulePayLoad) ToBytes() ([]byte, error) {
	content, err := json.Marshal(this)
	if nil != err {
		return nil, err
	}
	return content, nil
}

func (this *Trigger) SetSingle(single *Single) {
	this.Single_ = single
}

func (this *Trigger) SetPeriodical(periodical *Periodical) {
	this.Periodical_ = periodical
}

func FormatDateTime(t int64) string {
	str := time.Unix(t, 0).Format("2006-01-02 15:04:05")
	return str
}

func FormatDayTime(t int) string {
	hour := t / 3600
	monitue := (t % 3600) / 60
	second := t % 60
	str := fmt.Sprintf("%.2d:%.2d:%.2d", hour, monitue, second)
	return str
}

func (this *Single) SetTime(t int64) {
	this.Time = FormatDateTime(t)
}

func (this *Periodical) SetStartTime(t int64) {
	this.Start = FormatDateTime(t)
}

func (this *Periodical) SetEndTime(t int64) {
	this.End = FormatDateTime(t)
}

func (this *Periodical) SetTime(t int) {
	this.Time = FormatDayTime(t)
}

func (this *Periodical) SetTimeUnit(timeunit string) {
	this.TimeUnit = timeunit
}

func (this *Periodical) SetFrequency(frequency int) {
	this.Frequency = frequency
}

func (this *Periodical) AddPoint(point string) {
	this.Point = append(this.Point, point)
}
