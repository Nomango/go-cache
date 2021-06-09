package cache

import (
	"sync"
	"sync/atomic"
	"time"
)

// Item 缓存项
type Item struct {
	Value       interface{}
	ExpiredTime *time.Time
}

func NewItem(val interface{}, expiration time.Duration) *Item {
	if expiration == NoExpiration {
		return &Item{val, nil}
	}
	expiredTime := time.Now().Add(expiration)
	return &Item{
		Value:       val,
		ExpiredTime: &expiredTime,
	}
}

// IsExpired 对象是否过期
func (i *Item) IsExpired() bool {
	if i.ExpiredTime == nil {
		// 永不过期的对象
		return false
	}
	return time.Now().After(*i.ExpiredTime)
}

// ItemMap
type ItemMap interface {
	// GetItem 获取缓存项
	GetItem(key string) (*Item, bool)
	// AddItem 添加缓存项
	AddItem(key string, val *Item)
	// RemoveItem 移除缓存项
	RemoveItem(key string)
	// Flush 清空缓存
	Flush()
	// Len 返回缓存对象数量
	Len() int
	// Range 遍历缓存对象，接受一个op函数，函数参数分别是key/value
	// 返回true表示继续遍历，返回false表示停止遍历
	Range(op func(string, interface{}) bool)
	// ClearExpired 清空过期对象
	ClearExpired()
}

var _ ItemMap = &itemMap{}

type itemMap struct {
	items     atomic.Value // 实际是*sync.Map类型
	count     int64
	deletedCb DeletedCallback
}

func newItemMap(deletedCb DeletedCallback) ItemMap {
	m := &itemMap{}
	m.items.Store(&sync.Map{})
	m.deletedCb = deletedCb
	return m
}

func (m *itemMap) getItems() *sync.Map {
	// 保证读写*sync.Map是原子操作，否则执行Flush()会有并发问题
	return m.items.Load().(*sync.Map)
}

func (m *itemMap) GetItem(key string) (*Item, bool) {
	item, ok := m.getItems().Load(key)
	if ok {
		return item.(*Item), true
	}
	return nil, false
}

func (m *itemMap) AddItem(key string, val *Item) {
	_, ok := m.getItems().Load(key)
	if !ok {
		atomic.AddInt64(&m.count, 1)
	}
	m.getItems().Store(key, val)
}

func (m *itemMap) RemoveItem(key string) {
	val, ok := m.getItems().Load(key)
	if ok {
		m.remove(key, val.(*Item))
	}
}

func (m *itemMap) Flush() {
	if m.deletedCb != nil {
		// 逐个删除
		m.getItems().Range(func(key, val interface{}) bool {
			m.remove(key.(string), val.(*Item))
			return true
		})
		return
	}

	// 直接替换新的map
	m.items.Store(&sync.Map{})
	atomic.StoreInt64(&m.count, 0)
}

func (m *itemMap) Len() int {
	return int(atomic.LoadInt64(&m.count))
}

func (m *itemMap) Range(op func(string, interface{}) bool) {
	if op == nil {
		return
	}

	// sync.Map 的Range不会阻塞，可以放心执行
	m.getItems().Range(func(key, val interface{}) bool {
		item := val.(*Item)
		if item.IsExpired() {
			return true
		}
		if !op(key.(string), item.Value) {
			// break
			return false
		}
		return true
	})
}

func (m *itemMap) ClearExpired() {
	// sync.Map 的Range不会阻塞，可以放心执行
	m.getItems().Range(func(key, val interface{}) bool {
		item := val.(*Item)
		if item.IsExpired() {
			m.remove(key.(string), item)
		}
		return true
	})
}

func (m *itemMap) remove(key string, item *Item) {
	m.getItems().Delete(key)
	atomic.AddInt64(&m.count, -1)

	if m.deletedCb != nil {
		m.deletedCb(key, item.Value)
	}
}
