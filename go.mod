module github.com/jay-wlj/gobaselib

go 1.12

//replace github.com/jay-wlj/gobaselib => ./

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/disintegration/imaging v1.6.2 // indirect
	github.com/fatih/structs v1.1.0
	github.com/gin-gonic/gin v1.4.0
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/go-sql-driver/mysql v1.4.1
	github.com/gorilla/websocket v1.4.1
	github.com/jie123108/glog v0.0.0-20160701133742-ca74c069d4e1
	github.com/jie123108/imaging v1.1.0
	github.com/jinzhu/gorm v1.9.11
	github.com/json-iterator/go v1.1.7
	github.com/kr/pretty v0.1.0 // indirect
	github.com/labstack/echo v3.3.10+incompatible
	github.com/labstack/gommon v0.3.0 // indirect
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/levigross/grequests v0.0.0-20190908174114-253788527a1a
	github.com/nsqio/go-nsq v1.0.7
	//github.com/shopspring/decimal v0.0.0-20191009025716-f1972eb1d1f5
	github.com/shopspring/decimal v0.0.0-20180709203117-cd690d0c9e24 // 先用此版本,不会产生错误"pq: encode: unknown type for decimal.Decimal"
	github.com/valyala/fasttemplate v1.1.0 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible
	github.com/ziutek/mymysql v1.5.4 // indirect
	github.com/zwczou/jpush v0.0.0-20180527005611-a5e77e351698
	gopkg.in/go-playground/validator.v9 v9.30.0
	gopkg.in/gorp.v1 v1.7.2
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22
	gopkg.in/olahol/melody.v1 v1.0.0-20170518105555-d52139073376
	gopkg.in/redis.v5 v5.2.9
	gopkg.in/yaml.v2 v2.2.2
)
