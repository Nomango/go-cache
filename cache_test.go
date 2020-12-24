package cache_test

import (
	"math/rand"
	"runtime"
	"testing"
	"time"

	"github.com/Nomango/go-cache"
	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	c := cache.New()
	assert.Equal(t, c.Len(), 0)

	// 保存两个不过期的对象
	num1 := rand.Int63()
	c.Set("key1", num1)
	assert.Equal(t, c.Len(), 1)

	num2 := rand.Int63()
	c.Set("key2", num2)
	assert.Equal(t, c.Len(), 2)

	// 测试取出
	value, found := c.Get("key1")
	assert.Equal(t, found, true)
	assert.Equal(t, value, num1)

	value, found = c.Get("key2")
	assert.Equal(t, found, true)
	assert.Equal(t, value, num2)

	// 测试遍历
	c.Range(func(key string, value interface{}) bool {
		if key == "key1" {
			assert.Equal(t, value, num1)
			return true
		}
		if key == "key2" {
			assert.Equal(t, value, num2)
			return true
		}
		t.Errorf("Range key is unknown: %s", key)
		t.FailNow()
		return false
	})

	// 移除key2
	c.Delete("key2")
	value, found = c.Get("key2")
	assert.Equal(t, found, false)
	assert.Equal(t, value, nil)
	assert.Equal(t, c.Len(), 1)

	// 保存一个会过期的对象
	expiration := time.Second * 1
	c.SetWithExpiration("key3", rand.Int63(), expiration)
	assert.Equal(t, c.Len(), 2)

	startTime := time.Now()
	for {
		time.Sleep(time.Millisecond * 200)
		deltaTime := time.Now().Sub(startTime)
		// 手动清理缓存
		c.ClearExpired()
		if deltaTime < expiration {
			// 两个对象应该都未清掉
			assert.Equal(t, c.Len(), 2)
			continue
		}
		// 其中一个对象被清掉
		assert.Equal(t, c.Len(), 1)
		break
	}

	// 检查key3被清掉
	value, found = c.Get("key3")
	assert.Equal(t, found, false)
	assert.Equal(t, value, nil)

	// 检查key1未被清掉
	value, found = c.Get("key1")
	assert.Equal(t, found, true)
	assert.Equal(t, value, num1)

	// 清空缓存
	c.Flush()
	value, found = c.Get("key1")
	assert.Equal(t, found, false)
	assert.Equal(t, value, nil)
	assert.Equal(t, c.Len(), 0)
}

func TestCacheWithDefaultExpiration(t *testing.T) {
	expiration := time.Second * 1 // 默认1秒过期
	options := &cache.Options{
		DefaultExpiration: expiration,
	}
	c := cache.NewWithOptions(options)

	// 保存一个自动过期的对象
	num := rand.Int63()
	c.Set("key", num)

	startTime := time.Now()
	for {
		time.Sleep(time.Millisecond * 200)
		deltaTime := time.Now().Sub(startTime)
		// 手动清理缓存
		c.ClearExpired()
		if deltaTime < expiration {
			// 对象应该未清掉
			assert.Equal(t, c.Len(), 1)
			continue
		}
		// 对象被清掉
		assert.Equal(t, c.Len(), 0)
		break
	}
}

func TestCacheCleanUp(t *testing.T) {
	cleanInterval := time.Second * 1 // 每1秒自动清理一次
	options := &cache.Options{
		CleanInterval: cleanInterval,
	}
	c := cache.NewWithOptions(options)

	// 保存一个不过期的对象
	num1 := rand.Int63()
	c.SetWithExpiration("key1", num1, cache.NoExpiration /* 指定不过期 */)

	// 保存一个会过期的对象
	num2 := rand.Int63()
	expiration := time.Millisecond * 500
	c.SetWithExpiration("key2", num2, expiration)

	// 此时应该包含两个对象
	assert.Equal(t, c.Len(), 2)

	startTime := time.Now()
	for {
		time.Sleep(time.Millisecond * 200)
		deltaTime := time.Now().Sub(startTime)
		if deltaTime < cleanInterval {
			// 未达到自动清理时间
			continue
		}
		// 自动清理过后，应该剩下一个对象
		assert.Equal(t, c.Len(), 1)
		break
	}

	// 测试cleaner自动销毁
	c = nil
	runtime.GC()
	time.Sleep(time.Millisecond * 500)
}
