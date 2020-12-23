package cache_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/Nomango/go-cache"
	"github.com/stretchr/testify/assert"
)

func TestItem(t *testing.T) {
	num := rand.Int63()

	expiration := time.Second * 1
	item := cache.NewItem(num, expiration)

	startTime := time.Now()
	for {
		time.Sleep(time.Millisecond * 200)
		deltaTime := time.Now().Sub(startTime)
		if deltaTime < expiration {
			assert.Equal(t, item.IsExpired(), false)
			continue
		}
		assert.Equal(t, item.IsExpired(), true)
		break
	}

	assert.Equal(t, item.Value, num)
}
