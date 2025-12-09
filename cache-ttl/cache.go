package cachettl

import (
	"container/heap"
	"fmt"
	"sync"
	"time"
)

type Cache struct {
	rwMutex *sync.RWMutex
	heapq   *heapq
}

type CacheInterface interface {
	Set(key string, value any, ttl time.Duration)
	Get(key string) (any, bool)
}

func NewCacheInterface() CacheInterface {
	cache := Cache{
		rwMutex: &sync.RWMutex{},
		heapq: &heapq{
			pq:    PriorityQueue{},
			byKey: make(map[string]*Item),
		},
	}
	go cache.startCleanup(3 * time.Second)
	return &cache
}

// active go routine to clean the items that are expired: exclusive lock for writes
// startCleanup runs a ticker to remove expired items.
func (c *Cache) startCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	// ticker runs for every "interval" duration
	for range ticker.C {
		c.rwMutex.Lock()
		c.heapq.cleanPastTTL()
		c.rwMutex.Unlock()
	}
}

func (c *Cache) Set(key string, value any, ttl time.Duration) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()
	newExpiry := time.Now().Add(ttl)
	c.heapq.upsert(key, value, newExpiry)
}

func (c *Cache) Get(key string) (any, bool) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()
	return c.heapq.Get(key)
}

type heapq struct {
	pq    PriorityQueue
	byKey map[string]*Item
}

func (q *heapq) cleanPastTTL() {
	fmt.Println("cleaning process started in heapq : ...", q.pq.Len())
	for q.pq.Len() > 0 && q.pq[0].expiry.Before(time.Now()) {
		delete(q.byKey, q.pq[0].cache_key)
		heap.Pop(&q.pq)
	}
	fmt.Println("cleaning process ended in heapq : ...", q.pq.Len())

}

func (q *heapq) upsert(key string, value any, expiry time.Time) {
	// update the existing item
	if item, ok := q.byKey[key]; ok {
		item.cache_val = value
		item.expiry = expiry
		item.cache_key = key
		q.pq.update(item, key, expiry)
		return
	}

	// new item
	item := &Item{cache_val: value, cache_key: key, expiry: expiry}
	heap.Push(&q.pq, item)
	q.byKey[key] = item
}

func (q *heapq) Get(key string) (any, bool) {
	item, ok := q.byKey[key]
	// if not present or expired
	if !ok || item.expiry.Before(time.Now()) {
		return nil, false
	}
	return item.cache_val, true
}

// An Item is something we manage in a priority queue.
type Item struct {
	cache_val any
	cache_key string    // The value of the item; arbitrary.
	expiry    time.Time // The priority of the item in the queue.
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// min heap implementation based on ttl
	return pq[i].expiry.Before(pq[j].expiry)
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // don't stop the GC from reclaiming the item eventually
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// update modifies the priority and value of an Item in the queue.
func (pq *PriorityQueue) update(item *Item, key string, expiry time.Time) {
	heap.Fix(pq, item.index)
}
