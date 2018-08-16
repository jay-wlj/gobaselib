package jpushclient

type Notice struct {
	Alert    string          `json:"alert,omitempty"`
	Android  *AndroidNotice  `json:"android,omitempty"`
	IOS      *IOSNotice      `json:"ios,omitempty"`
	WINPhone *WinPhoneNotice `json:"winphone,omitempty"`
}

type AndroidNotice struct {
	Alert     string                 `json:"alert"`
	Title     string                 `json:"title,omitempty"`
	BuilderId int                    `json:"builder_id,omitempty"`
	Extras    map[string]interface{} `json:"extras,omitempty"`
}

type IOSNotice struct {
	Alert            map[string]string      `json:"alert"`
	Sound            string                 `json:"sound,omitempty"`
	Badge            int                    `json:"badge,omitempty"`
	ContentAvailable bool                   `json:"content-available,omitempty"`
	Category         string                 `json:"category,omitempty"`
	Extras           map[string]interface{} `json:"extras,omitempty"`
}

type WinPhoneNotice struct {
	Alert    string                 `json:"alert"`
	Title    string                 `json:"title,omitempty"`
	OpenPage string                 `json:"_open_page,omitempty"`
	Extras   map[string]interface{} `json:"extras,omitempty"`
}

func NewNotice(platforms []string, title string, content string, extras map[string]interface{}) (notice *Notice) {
	notice = &Notice{}
	notice.SetAlert(content)
	for _, platform_ := range platforms {
		if platform_ == "android" {
			androidnotice := AndroidNotice{Title: title, Alert: content, Extras: extras}
			notice.SetAndroidNotice(&androidnotice)
		} else if platform_ == "ios" {
			ios_alert := make(map[string]string)
			ios_alert["title"] = title
			ios_alert["body"] = content
			//alert := make(map[string]interface{})
			//alert["alert"] = ios_alert
			//bytes, _ := json.Marshal(alert)
			iosnotice := IOSNotice{Alert: ios_alert, Extras: extras, Badge: 0}
			notice.SetIOSNotice(&iosnotice)
		}
	}
	return notice
}

func (this *Notice) SetAlert(alert string) {
	this.Alert = alert
}

func (this *Notice) SetAndroidNotice(n *AndroidNotice) {
	this.Android = n
}

func (this *Notice) SetIOSNotice(n *IOSNotice) {
	this.IOS = n
}

func (this *Notice) SetWinPhoneNotice(n *WinPhoneNotice) {
	this.WINPhone = n
}
