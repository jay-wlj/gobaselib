package cache

import (
	"reflect"

	"github.com/jie123108/glog"
)

type StorageConfig struct {
	IndexCfg RedisConfig
	DataCfg  RedisConfig
	PageSize int
}

type GetIDCb func(interface{}) int64
type GetKeyCb func(int64) string
type GetIndexKeyCb func(interface{}) string
type MarshalCb func(v interface{}) ([]byte, error)
type UnmarshalCb func(data []byte, v interface{}) error

type StorageCallback struct {
	ID        GetIDCb       //资源的ID获取函数
	Key       GetKeyCb      //资源缓存使用的KEY获取函数.
	IndexKey  GetIndexKeyCb //资源的页缓存KEY获取函数.
	Marshal   MarshalCb     //将资源对象序列化成字节数组.
	Unmarshal UnmarshalCb   //将字节数据反序列化成对象
	Objtype   reflect.Type  //要存储的主对象的类型.
}

type RedisStorage struct {
	index *PageCache
	data  *RedisCache
	cfg   *StorageConfig
	cb    *StorageCallback
}

type ListData struct {
	Objs     []interface{} `json:"objs"`
	NextPage int           `json:"next_page"`
	CurPage  int           `json:"cur_page"`
	Total    int64         `json:"total"`
}

func NewRedisStorage(cfg *StorageConfig, cb *StorageCallback) (storage *RedisStorage, err error) {
	data, err := NewRedisCache(&cfg.DataCfg)
	if err != nil {
		glog.Errorf("NewRedisCache(%v) failed! err: %v", cfg.DataCfg, err)
		return nil, err
	}

	if cfg.PageSize < 1 {
		cfg.PageSize = 10
	}

	index, err := NewPageCache(&cfg.IndexCfg, cfg.PageSize)
	if err != nil {
		glog.Errorf("NewPageCache(%v, %v) failed! err: %v", cfg.IndexCfg, cfg.PageSize, err)
		return nil, err
	}

	storage = &RedisStorage{index, data, cfg, cb}
	return
}

/**
 * 修改一个对象
 * 只修改对象数据, 不修改索引数据
 */
func (this *RedisStorage) Update(obj interface{}) (err error) {
	id := this.cb.ID(obj)
	key := this.cb.Key(id)
	var val interface{}
	if this.cb.Marshal != nil {
		var buf []byte
		buf, err = this.cb.Marshal(obj)
		if err != nil {
			glog.Errorf("Marshal(%v) failed! err: %v", obj, err)
			return
		}
		val = string(buf)
	} else {
		val = obj
	}
	err = this.data.Set(key, val, 0)
	if err != nil {
		glog.Errorf("data.Set(%s, %v) failed! err: %v", key, obj, err)
		return
	}
	return
}

/**
 * 添加一个对象.
 */
func (this *RedisStorage) Add(obj interface{}) (err error) {
	id := this.cb.ID(obj)
	err = this.Update(obj)
	if err != nil {
		return
	}

	indexkey := this.cb.IndexKey(obj)

	_, err = this.index.HIncrBy(indexkey, "total", 1)
	if err != nil {
		glog.Errorf("index.HIncrBy(%s, 'total', 1) failed! err: %v", indexkey, err)
		return
	}

	err = this.index.Add(indexkey, id)
	if err != nil {
		glog.Errorf("index.Add(%s, %d) failed! err: %v", indexkey, id, err)
		return
	}
	return
}

/**
 * 删除一个对象
 */
func (this *RedisStorage) Del(indexkey string, id int64) (n int64, err error) {
	n, err = this.data.Del(this.cb.Key(id))
	if n > 0 {
		_, err = this.index.HIncrBy(indexkey, "total", -1)
		if err != nil {
			glog.Errorf("index.HIncrBy(%s, 'total', -1) failed! err: %v", indexkey, err)
			return
		}
	}
	return
}

/**
 * 获取总数.
 */
func (this *RedisStorage) GetTotal(indexkey string) (total int64, err error) {
	total, err = this.index.HGetI(indexkey, "total")
	if err != nil {
		glog.Errorf("index.HGetI(%s, 'total') failed! err: %v", indexkey, err)
		return
	}
	return
}

/**
 * 删除一个资源的所有索引
 */
func (this *RedisStorage) DelAllIndex(indexkey string) (n int64, err error) {
	return this.index.DelAll(indexkey)
}

/**
 * 查询一页数据
 * indexkey索引页的主键
 * page 是查询的页码, 0表示查询最新的一页.
 * 返回值中,
 */
func (this *RedisStorage) List(indexkey string, page int) (data ListData, err error) {
	pageDetail, err := this.index.Get(indexkey, page)
	//当没有下一页时, next_page返回-1. 如果返回0, 客户端处理不当, 会变成死循环.
	data.NextPage = -1
	if err != nil {
		if err.Error() != "redis: nil" {
			glog.Errorf("index.Get(%s, %d) failed! err: %v", indexkey, page, err)
		}
		return data, err
	}

	total, err := this.GetTotal(indexkey)
	if err != nil {
		glog.Errorf("GetTotal(%s) failed! err: %v", indexkey, err)
		return data, err
	}
	data.Total = total

	page = pageDetail.CurPage
	data.CurPage = page
	data.Objs = make([]interface{}, len(pageDetail.Data))
	realsize := 0
	for i := 0; i < len(pageDetail.Data); i++ {
		id := pageDetail.Data[i]
		key := this.cb.Key(id)
		buf, err := this.data.GetB(key)
		// 不存在, 删除指定数据.
		if err == ErrNotExist {
			err = this.index.Del(indexkey, page, id)
			if err != nil {
				glog.Errorf("index.Del(%s, page:%d, id:%d) failed!", indexkey, page, id)
			}
			continue
		} else if err != nil {
			data.Objs = nil
			glog.Errorf("index.Del(%s, page:%d, id:%d) failed! err:%v", indexkey, page, id, err)
			return data, err
		}
		val := reflect.New(this.cb.Objtype).Interface()
		err = this.cb.Unmarshal(buf, val)
		if err != nil {
			glog.Errorf("%s, page:%d, Unmarshal(%v), failed! err: %v", indexkey, page, buf, err)
			return data, err
		}
		data.Objs[realsize] = val
		realsize += 1
	}
	if realsize < len(pageDetail.Data) {
		data.Objs = data.Objs[0:realsize]
	}

	if len(data.Objs) == 0 && pageDetail.Pre > 0 && pageDetail.Pre != page {
		page = pageDetail.Pre
		return this.List(indexkey, page)
	}

	/* 注意, 因为pagecache中存储时, 是从旧到新存储数据的,
	   但是查询列表是从新到旧取的,所以next_page为Pre的值
	*/
	data.NextPage = pageDetail.Pre
	if data.NextPage == 0 {
		data.NextPage = -1
	}
	return data, err
}

func (this *RedisStorage) Get(id int64) (data interface{}, err error) {
	if id <= 0 {
		return
	}
	key := this.cb.Key(id)
	buf, err := this.data.GetB(key)
	// 不存在, 删除指定数据.
	if err == ErrNotExist {
		return
	} else if err != nil {
		return
	}
	val := reflect.New(this.cb.Objtype).Interface()
	err = this.cb.Unmarshal(buf, val)

	return val, err
}
