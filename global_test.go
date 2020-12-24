package cache_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/Nomango/go-cache"
	"github.com/stretchr/testify/assert"
)

func TestGlobalCache(t *testing.T) {
	// 测试 lazy init
	// cache.Init(nil)

	// 获取全局缓存
	global := cache.Global()
	assert.Equal(t, global.Len(), 0)

	// 保存一个不过期的对象
	num1 := rand.Int63()
	cache.Set("key1", num1)
	assert.Equal(t, global.Len(), 1)

	// 测试取出
	value, found := cache.Get("key1")
	assert.Equal(t, found, true)
	assert.Equal(t, value, num1)

	// 保存一个会过期的对象
	expiration := time.Second * 1
	cache.SetWithExpiration("key3", rand.Int63(), expiration)
	assert.Equal(t, global.Len(), 2)

	startTime := time.Now()
	for {
		time.Sleep(time.Millisecond * 200)
		deltaTime := time.Now().Sub(startTime)
		// 手动清理缓存
		global.ClearExpired()
		if deltaTime < expiration {
			// 两个对象应该都未清掉
			assert.Equal(t, global.Len(), 2)
			continue
		}
		// 其中一个对象被清掉
		assert.Equal(t, global.Len(), 1)
		break
	}

	// 检查key3被清掉
	value, found = cache.Get("key3")
	assert.Equal(t, found, false)
	assert.Equal(t, value, nil)

	// 检查key1未被清掉
	value, found = cache.Get("key1")
	assert.Equal(t, found, true)
	assert.Equal(t, value, num1)
}
