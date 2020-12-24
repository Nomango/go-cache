package cache_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/Nomango/go-cache"
	"github.com/stretchr/testify/assert"
)

func TestLRUCache(t *testing.T) {
	options := &cache.Options{
		Capacity: 2,
	}
	c := cache.NewWithOptions(options)
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

func TestLRUCacheCapacity(t *testing.T) {
	options := &cache.Options{
		Capacity: 2, // 容量为2
	}
	c := cache.NewWithOptions(options)

	c.Set("key1", 1)
	c.Set("key2", 2)
	assert.Equal(t, c.Len(), 2)

	// 超过容量，自动清除最后一个
	c.Set("key3", 3)
	assert.Equal(t, c.Len(), 2)
	_, found := c.Get("key1")
	assert.Equal(t, found, false)
	_, found = c.Get("key2")
	assert.Equal(t, found, true)
	_, found = c.Get("key3")
	assert.Equal(t, found, true)

	// 使用key2，然后新加key4，应自动清除key3
	_, _ = c.Get("key2")
	c.Set("key4", 4)

	assert.Equal(t, c.Len(), 2)
	_, found = c.Get("key3")
	assert.Equal(t, found, false)
	_, found = c.Get("key2")
	assert.Equal(t, found, true)
	_, found = c.Get("key4")
	assert.Equal(t, found, true)

	// 覆盖保存key2，然后新加key5，应自动清除key4
	c.Set("key2", -1)
	c.Set("key5", 5)

	assert.Equal(t, c.Len(), 2)
	_, found = c.Get("key4")
	assert.Equal(t, found, false)
	_, found = c.Get("key5")
	assert.Equal(t, found, true)
	key2Val, found := c.Get("key2")
	assert.Equal(t, found, true)
	assert.Equal(t, key2Val, -1) // key2的值正确覆盖
}
