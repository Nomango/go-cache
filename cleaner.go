package cache

import (
	"time"
)

type cleaner struct {
	interval    time.Duration
	stopEvicter chan bool
}

// cacheWapper 包装器，为了正确执行finalizer而使用
type cacheWapper struct {
	Cache
	cleaner *cleaner
}

var _ Cache = &cacheWapper{}

func newCleaner(cache *cache, interval time.Duration) *cleaner {
	c := &cleaner{
		interval:    interval,
		stopEvicter: make(chan bool),
	}
	go c.Run(cache)
	return c
}

func (c *cleaner) Run(cache *cache) {
	t := time.NewTicker(c.interval)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			cache.ClearExpired()
		case <-c.stopEvicter:
			return
		}
	}
}

func (c *cleaner) Stop() {
	close(c.stopEvicter)
}

func cacheFinalizer(c *cacheWapper) {
	c.cleaner.Stop()
}
