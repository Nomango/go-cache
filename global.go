package cache

import (
	"sync"
	"time"
)

type globalCache struct {
	cache Cache
	once  sync.Once
}

func (g *globalCache) lazyInit(options *Options) {
	g.once.Do(func() {
		g.cache = NewWithOptions(options)
	})
}

var global globalCache

// Init 初始化全局缓存
func Init(options *Options) {
	global.lazyInit(options)
}

// Set 缓存一个对象
func Set(key string, val interface{}) {
	global.lazyInit(nil)
	global.cache.Set(key, val)
}

// SetWithExpiration 缓存一个对象，并设置过期时间
func SetWithExpiration(key string, val interface{}, expiration time.Duration) {
	global.lazyInit(nil)
	global.cache.SetWithExpiration(key, val, expiration)
}

// Get 获取一个缓存对象
func Get(key string) (value interface{}, found bool) {
	global.lazyInit(nil)
	return global.cache.Get(key)
}

// Delete 删除一个缓存对象
func Delete(key string) {
	global.lazyInit(nil)
	global.cache.Delete(key)
}

// Global 获取全局缓存
func Global() Cache {
	global.lazyInit(nil)
	return global.cache
}
