package cache

import (
	"container/list"
	"sync"
)

type lruItemMap struct {
	items map[string]*list.Element
	mu    sync.RWMutex
	// LRU缓存的容量
	capacity int
	// LRU链表
	list *list.List
}

// lruNode 链表节点
type lruNode struct {
	key  string
	item *Item
}

func newLRUItemMap(capacity int) ItemMap {
	return &lruItemMap{
		items:    make(map[string]*list.Element, capacity),
		capacity: capacity,
		list:     list.New(),
	}
}

func (m *lruItemMap) GetItem(key string) (*Item, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	elem, ok := m.items[key]
	if ok {
		// 将新访问的元素放到链表头
		m.list.Remove(elem)
		newElem := m.list.PushFront(elem.Value)
		// 更新map中的elem
		m.items[key] = newElem
		return newElem.Value.(*lruNode).item, ok
	}
	return nil, false
}

func (m *lruItemMap) AddItem(key string, val *Item) {
	m.mu.Lock()
	defer m.mu.Unlock()

	oldElem, ok := m.items[key]

	// 保存新节点
	newNode := &lruNode{key: key, item: val}
	elem := m.list.PushFront(newNode)
	m.items[key] = elem

	// 已经存在key，直接覆盖
	if ok {
		m.list.Remove(oldElem)
		return
	}

	// 不存在key，且超过容量
	size := len(m.items)
	if size > m.capacity {
		// 移除最后一个
		back := m.list.Back()
		m.list.Remove(back)
		delete(m.items, back.Value.(*lruNode).key)
	}
}

func (m *lruItemMap) RemoveItem(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	elem, ok := m.items[key]
	if ok {
		m.list.Remove(elem)
		delete(m.items, key)
	}
}

func (m *lruItemMap) Flush() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items = make(map[string]*list.Element)
	m.list = list.New()
}

func (m *lruItemMap) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.list.Len()
}

func (m *lruItemMap) Range(op func(string, interface{}) bool) {
	if op == nil {
		return
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	for elem := m.list.Front(); elem != nil; elem = elem.Next() {
		node := elem.Value.(*lruNode)
		if node.item.IsExpired() {
			continue
		}
		if !op(node.key, node.item.Value) {
			break
		}
	}
}

func (m *lruItemMap) ClearExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for key, elem := range m.items {
		node := elem.Value.(*lruNode)
		if node.item.IsExpired() {
			m.list.Remove(elem)
			delete(m.items, key)
			count++
		}

		// Benchmark测试该方法的性能如下
		// |待删除的对象数量|一次CleanUp耗时|
		// |-------------|--------------|
		// |1k           |697.929µs     |
		// |1w           |4.749823ms    |
		// |10w          |53.20541ms    |
		// |100w         |806.008161ms  |
		// 当删除对象数量不超过1w时，一次清理操作耗时<10ms，可以做到用户无感知
		// LRU限制了对象容量，且cache.Get方法也会清理过期对象，故不需要担心清理不及时导致的内存问题
		if count > 10000 {
			break
		}
	}
}
