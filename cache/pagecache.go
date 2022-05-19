package cache

import (
	"gobaselib/log"
	"strconv"
	// "encoding/json"

	"github.com/vmihailenco/msgpack"
)

type PageCache struct {
	Config   *RedisConfig
	PageSize int
	Cache    *RedisCache
}

func NewPageCache(cfg *RedisConfig, page_size int) (cache *PageCache, err error) {
	Cache, err := NewRedisCache(cfg)
	if err != nil {
		return nil, err
	}

	cache = &PageCache{cfg, page_size, Cache}

	return
}

func (this *PageCache) Clean(key string) (n int64, err error) {
	n, err = this.Cache.Del(key)
	return
}

var SCRIPT_ADD = `
    local key = KEYS[1]
    local value = tonumber(ARGV[1])
    local page_size = tonumber(ARGV[2])
    local cur_page = redis.call('hget', key, 'cur_page')
    if not cur_page then
        redis.call('hset', key, 'cur_page', 1)
        cur_page = 1
    else
        cur_page = tonumber(cur_page)
    end
    -- local pagekey = tostring(cur_page)
    local str = redis.call('hget', key, cur_page)

    -- decode str
    local vals = nil
    if str then
        vals = cmsgpack.unpack(str)
        assert(type(vals) == 'table', 'decode(' .. str .. ") failed!")
    else
        vals = {d={}, n=2}
    end

    -- 去重逻辑, 相邻的两个元素, 不允许重复.
    if #vals.d > 0 and vals.d[1] == value then
        return 1
    end

    if #vals.d >= page_size then
        local new_page = tostring(redis.call('hincrby', key, 'cur_page', 1))
        local new_node = cmsgpack.pack({p=cur_page,d={value}, n=new_page+1})
        redis.call('hset', key, new_page, new_node)
    else
        table.insert(vals.d, 1, value)
        local valstr = cmsgpack.pack(vals)
        -- redis.log(redis.LOG_WARNING, "vals: ", valstr, cjson.encode(vals))
        redis.call('hset', key, cur_page, valstr)
    end
    return 1
`

type PageDetail struct {
	Next    int     `msgpack:"n"`
	Pre     int     `msgpack:"p"`
	Data    []int64 `msgpack:"d"`
	CurPage int     `msgpack:"-"` //当查询最新一页时, 返回当前页号
}

/** 结构定义
// page_size = 5
redis:
key: RedisHash{
    "cur_page": 3,
    "count": 20,
    "1": {p=nil, d=[5,4,3,2,1], n=2},
    "2": {p=1, d=[10,9,8,7,6], n=3},
    "3": {p=2, d=[12,11], n=nil},
}
// cur_page 当前最新的页码.
**/
func (this *PageCache) Add(key string, value int64) (err error) {
	keys := []string{key}
	_, err = this.Cache.Eval(SCRIPT_ADD, keys, value, this.PageSize)
	return
}

func (this *PageCache) HSetI(key string, field string, val int64) (err error) {
	err = this.Cache.HSet(key, field, val, 0)
	return
}

func (this *PageCache) HIncrBy(key string, field string, incr int64) (n int64, err error) {
	n, err = this.Cache.HIncrBy(key, field, incr)
	return
}

func (this *PageCache) HGetI(key string, field string) (val int64, err error) {
	val, err = this.Cache.HGetI(key, field)
	return
}

// 删除某一个Key的所有数据.
func (this *PageCache) DelAll(keys ...string) (n int64, err error) {
	n, err = this.Cache.Del(keys...)
	return
}

var SCRIPT_DEL_PAGE = ` -- line 0
    local key = KEYS[1]
    local page = ARGV[1]
    local pagekey = tostring(page)
    local str = redis.call('hget', key, pagekey)

    -- 当前页数据不存在
    if str == false or str == nil then
        return 1
    end
    local cur_page = cmsgpack.unpack(str)
    if cur_page == nil then
        error(tostring(key).."[" .. pagekey .. "]'s value is invalid!")
    end

    local cur_page_no = redis.call('hget', key, 'cur_page')
    -- 要删除的是最新的一页, 需要同时修改cur_page值.
    if cur_page_no == page then
        local pre_page_no = cur_page.p
        redis.call('hset', key, 'cur_page', pre_page_no)
    end

    if cur_page.n then
        local pagekey = tostring(cur_page.n)
        local next_page_str = redis.call('hget', key, pagekey)
        if next_page_str then
            local next_page = cmsgpack.unpack(next_page_str)
            if next_page and type(next_page) == 'table' then
                next_page.p = cur_page.p
                redis.call('hset', key, pagekey, cmsgpack.pack(next_page))
            else
                error("key:" .. key .. ", pagekey:" .. pagekey .. " cmsgpack.unpack(" .. next_page_str .. ") failed!")
            end
        end
    end

    if cur_page.p then
        local pagekey = tostring(cur_page.p)
        local pre_page_str = redis.call('hget', key, pagekey)
        if pre_page_str then
            local pre_page = cmsgpack.unpack(pre_page_str)
            if pre_page and type(pre_page) == 'table' then
                pre_page.n = cur_page.n
                redis.call('hset', key, pagekey, cmsgpack.pack(pre_page))
            else
                error("key:" .. key .. ", pagekey:" .. pagekey .. " cmsgpack.unpack(" .. next_page_str .. ") failed!")
            end
        end
    end
    redis.call('hdel', key, pagekey)
    return 1
`

func (this *PageCache) DelPage(key string, page int) (err error) {
	log.Infof("del page key:%s, page: %d", key, page)
	keys := []string{key}
	_, err = this.Cache.Eval(SCRIPT_DEL_PAGE, keys, page)
	return err
}

var SCRIPT_DEL = `
    local key = KEYS[1]
    local page = ARGV[1]
    local value = tonumber(ARGV[2])
    local pagekey = tostring(page)
    --redis.log(redis.LOG_WARNING, "hget ", key, ", ", pagekey)
    local str = redis.call('hget', key, pagekey)
    if not str then
        return 1
    end
    --redis.log(redis.LOG_WARNING, "hget ", key, pagekey, ":" .. tostring(str))

    -- decode str
    local vals= cmsgpack.unpack(str)
    if type(vals) == 'table' and type(vals.d) == 'table' then
        for i = #vals.d, 1, -1 do
            if vals.d[i] == value then
                table.remove(vals.d, i)
                break
            end
        end
        -- 判断 #vals.d == 0
        local empty = #vals.d == 0
        -- redis.log(redis.LOG_WARNING, "vals.d: ", #vals.d, " empty:", tostring(empty))
        -- 保存
        local valstr = cmsgpack.pack(vals)

        redis.call('hset', key, pagekey, valstr)
        if empty then
            return 2
        end
        return 1
    else
        error("hget('" .. key .. "," .. pagekey .. ") data:[" .. str .. "] invalid!")
    end
    return 1
`

// 删除某一页中的一条记录.
func (this *PageCache) Del(key string, page int, value int64) (err error) {
	keys := []string{key}
	n, err := this.Cache.Eval(SCRIPT_DEL, keys, page, value)
	// log.Errorf("cache.Del(%s,%d,%d) n: %d", key, page, value, n)
	if n == 2 {
		return this.DelPage(key, page)
	}

	return err
}

// 如果key不存在, 返回 true, nil
func (this *PageCache) get_last_page(key string) (cur_page int, err error) {
	// 获取当前页(TODO: 添加缓存)
	s_cur_page, err := this.Cache.HGet(key, "cur_page")
	if err != nil {
		if err.Error() != "redis: nil" {
			log.Errorf("cache:hget(%s,'cur_page') failed! err: [%v]", key, err)
		}
		return
	}
	if s_cur_page == "" {
		return
	}

	cur_page, err = strconv.Atoi(s_cur_page)
	return
}

func (this *PageCache) Get(key string, page int) (pageDetail *PageDetail, err error) {
	if page == 0 {
		page, err = this.get_last_page(key)
		if err != nil {
			return nil, err
		}
		if page == 0 {
			return nil, err
		}
	}

	pagekey := strconv.Itoa(page)
	str, err := this.Cache.HGetB(key, pagekey)
	if err != nil {
		//log.Errorf("cache:hget(%s, %s) failed! err: %v", key, pagekey, err)
		return nil, err
	}

	if str == nil {
		return nil, ErrNotExist
	}

	// decode str
	pageDetail = &PageDetail{}
	err = msgpack.Unmarshal(str, pageDetail)
	if err != nil {
		log.Errorf("hget(%s,%s): msgpack.Unmarshal(%v) failed! err: %v", key, pagekey, str, err)
		return nil, err
	}
	pageDetail.CurPage = page

	return
}
