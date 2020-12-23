package cache

import (
	"runtime"
	"time"
)

const (
	// NoExpiration 永不过期
	NoExpiration time.Duration = 0
	// DefaultCleanInterval 默认的清空缓存时长
	DefaultCleanInterval time.Duration = time.Minute
)

// Cache 缓存器
type Cache interface {
	// Set 缓存一个对象
	Set(key string, val interface{})
	// SetWithExpiration 缓存一个对象，并设置过期时间
	SetWithExpiration(key string, val interface{}, expiration time.Duration)
	// Get 获取一个缓存对象
	Get(key string) (value interface{}, found bool)
	// 实现ItemMap接口的所有方法
	ItemMap
}

// Options 缓存选项
type Options struct {
	DefaultExpiration time.Duration
	CleanInterval     time.Duration
	Capacity          int
}

// New 新建缓存器
func New() Cache {
	return NewWithOptions(&Options{})
}

// NewWithOptions 新建缓存器
func NewWithOptions(options *Options) Cache {
	var m ItemMap
	if options.Capacity <= 0 {
		// 无容量上限的缓存
		m = newItemMap()
	} else {
		// LRU缓存
		m = newLRUItemMap(options.Capacity)
	}

	c := &cache{
		ItemMap: m,
		options: options,
	}
	if options != nil && options.CleanInterval > 0 {
		// 启动cleaner协程
		cleaner := newCleaner(c, options.CleanInterval)
		// 创建包装器
		wapper := &cacheWapper{c, cleaner}
		runtime.SetFinalizer(wapper, cacheFinalizer)
		return wapper
	}
	return c
}

var _ Cache = &cache{}

// cache 缓存器，不暴露给外部使用
type cache struct {
	ItemMap
	options *Options
}

func (c *cache) Set(key string, val interface{}) {
	expiration := NoExpiration
	if c.options != nil {
		expiration = c.options.DefaultExpiration
	}
	c.SetWithExpiration(key, val, expiration)
}

func (c *cache) SetWithExpiration(key string, val interface{}, expiration time.Duration) {
	c.AddItem(key, NewItem(val, expiration))
}

func (c *cache) Get(key string) (value interface{}, found bool) {
	item, ok := c.GetItem(key)
	if !ok {
		return nil, false
	}
	if item.IsExpired() {
		c.RemoveItem(key)
		return nil, false
	}
	return item.Value, true
}
