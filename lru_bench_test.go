package cache_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/Nomango/go-cache"
)

func BenchmarkLRUCache(b *testing.B) {
	// 测试cache.Set性能
	options := &cache.Options{
		Capacity: 10000, // 容量为1w
	}
	c := cache.NewWithOptions(options)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c.Set(fmt.Sprintf("%d", i), i)
	}
}

func BenchmarkLRUCacheConcurrent(b *testing.B) {
	// 测试cache.Set cache.Get并发
	options := &cache.Options{
		Capacity: 10000, // 容量为1w
	}
	c := cache.NewWithOptions(options)

	for i := 0; i < 10000; i++ {
		c.Set(fmt.Sprintf("%d", i), i)
	}

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("%d", i)
			c.Set(key, i)
			c.Get(key)
			i++
		}
	})
}

func BenchmarkLRUCacheCleanUp(b *testing.B) {
	// 测试cache.CleanUp性能
	testFunc := func(b *testing.B, capacity int) {
		options := &cache.Options{
			Capacity: capacity,
		}
		c := cache.NewWithOptions(options)

		for i := 0; i < b.N; i++ {
			b.StopTimer()
			for i := 0; i < capacity; i++ {
				c.SetWithExpiration(fmt.Sprintf("%d", i), i, time.Millisecond)
			}
			b.StartTimer()
			c.ClearExpired()
		}
	}

	initCapacity := 1000
	for i := 0; i < 4; i++ {
		capacity := initCapacity
		b.Run(fmt.Sprintf("Capacity %d", capacity), func(b *testing.B) {
			testFunc(b, capacity)
		})
		initCapacity *= 10
	}
}
