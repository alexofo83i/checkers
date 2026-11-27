package main

import (
	"container/list"
	"log"
	"sync"
	"time"
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

// clear сбрасывает весь кэш за O(1).
// Старые map/list отдаются GC целиком, без обхода элементов.
// make(...) с capacity сразу создаёт bucket'ы — это избавляет от рехешей
// при последующем массовом put (что и есть типичный сценарий вызова clear).
func (c *lruCache) clear() {
	c.mx.Lock()
	defer c.mx.Unlock()

	c.items = make(map[int64]*list.Element, c.capacity)
	c.order = list.New()
}

// bulkPut — вариант put для батча: один Lock на всю пачку.
// Принимает уже упакованные *CompactState — упаковка на вызывающей стороне.
func (c *lruCache) bulkPut(states []*CompactState) {
	c.mx.Lock()
	defer c.mx.Unlock()

	for _, cs := range states {
		if cs == nil {
			continue
		}
		key := cs.ID
		val := cs

		if el, ok := c.items[key]; ok {
			el.Value.(*playEntry).value = val
			c.order.MoveToFront(el)
			continue
		}
		el := c.order.PushFront(&playEntry{key: key, value: val})
		c.items[key] = el

		if c.order.Len() > c.capacity {
			oldest := c.order.Back()
			if oldest != nil {
				c.order.Remove(oldest)
				delete(c.items, oldest.Value.(*playEntry).key)
			}
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

// StoreBatch добавляет коллекцию состояний в кэш, предварительно (опционально)
// очищая его. Логирует количество элементов и время выполнения.
func (playstore *PlayStore) StoreBatch(states []*PlayState, clearFirst bool) {
	if clearFirst {
		playstore.cache.clear()
		log.Printf("playstore: cache cleared before batch insert")
	}

	total := len(states)
	if total == 0 {
		log.Printf("playstore: StoreBatch called with 0 elements")
		return
	}

	if total > playstore.cache.capacity {
		log.Printf("playstore: batch size %d exceeds capacity %d — old entries will be evicted",
			total, playstore.cache.capacity)
	}

	start := time.Now()

	// Упаковываем на вызывающей стороне, чтобы bulkPut принимал готовые *CompactState.
	compacted := make([]*CompactState, 0, total)
	for _, ps := range states {
		if ps == nil {
			continue
		}
		compacted = append(compacted, ps.packState())
	}

	playstore.cache.bulkPut(compacted)

	log.Printf(
		"playstore: stored %d/%d elements in cache in %s (capacity=%d, now holds=%d)",
		len(compacted), total, time.Since(start), playstore.cache.capacity, playstore.Len(),
	)
}

// All возвращает срез всех PlayState (распакованных) в порядке MRU→LRU.
// ВНИМАНИЕ: для больших кэшей это дорого (O(n) + unpack каждого).
// Указатели копируются под RLock, распаковка выполняется уже после снятия локалки.
func (playstore *PlayStore) All() []*PlayState {
	playstore.cache.mx.RLock()
	raw := make([]*CompactState, 0, playstore.cache.order.Len())
	for el := playstore.cache.order.Front(); el != nil; el = el.Next() {
		raw = append(raw, el.Value.(*playEntry).value)
	}
	playstore.cache.mx.RUnlock()

	out := make([]*PlayState, 0, len(raw))
	for _, cs := range raw {
		if cs != nil {
			out = append(out, cs.unpackState())
		}
	}
	return out
}

// AllCompact возвращает срез *CompactState без распаковки —
// удобно, когда нужен только обзор/дамп/сериализация.
func (playstore *PlayStore) AllCompact() []*CompactState {
	playstore.cache.mx.RLock()
	defer playstore.cache.mx.RUnlock()

	out := make([]*CompactState, 0, playstore.cache.order.Len())
	for el := playstore.cache.order.Front(); el != nil; el = el.Next() {
		out = append(out, el.Value.(*playEntry).value)
	}
	return out
}

// Len возвращает текущее количество элементов (для логирования).
func (playstore *PlayStore) Len() int {
	playstore.cache.mx.RLock()
	defer playstore.cache.mx.RUnlock()
	return playstore.cache.order.Len()
}

// func hash(s string) uint32 {
// 	h := fnv.New32a()
// 	h.Write([]byte(s))
// 	return h.Sum32()
// }

func (playState *PlayState) Hashcode() int64 {
	return playState.getCachedID()
}
