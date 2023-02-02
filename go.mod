module github.com/jay-wlj/gobaselib

go 1.12

//replace github.com/jay-wlj/gobaselib => ./

require (
	github.com/brianvoe/gofakeit/v6 v6.20.1
	github.com/disintegration/imaging v1.6.2 // indirect
	github.com/fatih/structs v1.1.0
	github.com/gin-gonic/gin v1.6.3
	github.com/go-redis/redis/v8 v8.11.5
	github.com/golang/snappy v0.0.4 // indirect
	github.com/jie123108/imaging v1.1.0
	github.com/json-iterator/go v1.1.9
	github.com/levigross/grequests v0.0.0-20190908174114-253788527a1a
	github.com/mitchellh/go-homedir v1.1.0
	github.com/nsqio/go-nsq v1.0.7
	github.com/onsi/gomega v1.19.0 // indirect
	github.com/pkg/sftp v1.12.0
	//github.com/shopspring/decimal v0.0.0-20191009025716-f1972eb1d1f5
	github.com/shopspring/decimal v0.0.0-20180709203117-cd690d0c9e24 // 先用此版本,不会产生错误"pq: encode: unknown type for decimal.Decimal"
	github.com/sirupsen/logrus v1.2.0
	github.com/vmihailenco/msgpack v4.0.4+incompatible
	github.com/zwczou/jpush v0.0.0-20180527005611-a5e77e351698
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v9 v9.30.0
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22
	gopkg.in/redis.v5 v5.2.9
	gopkg.in/yaml.v2 v2.4.0
	gorm.io/driver/mysql v1.4.5
	gorm.io/gorm v1.24.3
)
