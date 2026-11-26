package main

import (
	"container/list"
	"sync"
)

// entry holds the cache key/value and a pointer to the list element,
// so we can move/remove it in O(1).
type playEntry struct {
	key   int64
	value *CompactState
}

// lruCache is a fixed-size LRU cache with a mutex.
// When full, the least-recently-used item is evicted automatically.
type lruCache struct {
	mx       sync.RWMutex
	capacity int
	items    map[int64]*list.Element
	order    *list.List // front = most recently used
}

func newLRU(capacity int) *lruCache {
	return &lruCache{
		capacity: capacity,
		items:    make(map[int64]*list.Element, capacity),
		order:    list.New(),
	}
}

// get returns the value and moves it to the front (MRU).
func (c *lruCache) get(key int64) (*CompactState, bool) {
	c.mx.Lock()
	defer c.mx.Unlock()

	if el, ok := c.items[key]; ok {
		c.order.MoveToFront(el)
		return el.Value.(*playEntry).value, true
	}
	return nil, false
}

// put inserts or updates an entry; evicts the LRU entry if over capacity.
func (c *lruCache) put(key int64, value *CompactState) {
	c.mx.Lock()
	defer c.mx.Unlock()

	if el, ok := c.items[key]; ok {
		el.Value.(*playEntry).value = value
		c.order.MoveToFront(el)
		return
	}

	el := c.order.PushFront(&playEntry{key: key, value: value})
	c.items[key] = el

	if c.order.Len() > c.capacity {
		oldest := c.order.Back()
		if oldest != nil {
			c.order.Remove(oldest)
			delete(c.items, oldest.Value.(*playEntry).key)
		}
	}
}

type PlayStore struct {
	cache *lruCache
}

const playStoreCapacity = 1000000 // max entries kept in memory

var playStore = new(PlayStore).init()

func (playstore *PlayStore) init() *PlayStore {
	playstore.cache = newLRU(playStoreCapacity)
	return playstore
}

func (playstore *PlayStore) Get(key int64) (*PlayState, bool) {
	if v, ok := playstore.cache.get(key); ok && v != nil {
		return v.unpackState(), true
	}
	// if not found -> load from Postgresql database here
	return nil, false
}

func (playstore *PlayStore) Store(playState *PlayState) {
	playstore.cache.put(playState.Hashcode(), playState.packState())
}

// func hash(s string) uint32 {
// 	h := fnv.New32a()
// 	h.Write([]byte(s))
// 	return h.Sum32()
// }

func (playState *PlayState) Hashcode() int64 {
	return playState.getCachedID()
}
