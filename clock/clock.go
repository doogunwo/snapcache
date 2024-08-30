package clock

import "container/ring"
import "sync"

type ClockCache struct {
	maxEntries uint
	clockCount uint8
	hand       *ring.Ring
	cache      map[interface{}]*ring.Ring
	mutex      sync.RWMutex
	onEvict    func(interface{}, interface{})
}

type entry struct {
	AccessCounter uint8
	Key           interface{}
	Value         interface{}
}

// New creates a clock cache instance. The maxEntries param indicates the maximum number of items the cache can hold.
// If this number is exceeded an eviction occurs.
// The function takes an optional list of parameters that further customizes the instance.
func New(maxEntries uint, options ...Option) *ClockCache {
	r := ring.New(int(maxEntries))
	r.Value = &entry{}
	for p := r.Next(); p != r; p = p.Next() {
		p.Value = &entry{}
	}
	c := &ClockCache{
		maxEntries: maxEntries,
		clockCount: 1,
		hand:       r,
		cache:      make(map[interface{}]*ring.Ring),
		onEvict:    nil}

	for _, opt := range options {
		opt(c)
	}

	return c
}

type Option func(*ClockCache)

// Count is a parameter for the New function that sets the 'clock count' of the generalized clock replacement algorithm.
// For more information see http://www-inst.eecs.berkeley.edu/~cs266/sp10/readings/smith78.pdf page 11.
// If this option is not provided, cache instances will assume a default clock count value of 1.
func Count(count uint8) Option {
	return func(c *ClockCache) {
		c.clockCount = count
	}
}

// OnEvict registers a callback function for evictions.
// The function takes as arguments two interface{}'s that represents key and value of the evicted object respectivelly.
func OnEvict(cb func(interface{}, interface{})) Option {
	return func(c *ClockCache) {
		c.onEvict = cb
	}
}

// Put inserts a new entry key-value pair into the cache
func (c *ClockCache) Put(key interface{}, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if e, ok := c.cache[key]; ok {
		e.Value.(*entry).AccessCounter = c.clockCount
		e.Value.(*entry).Value = value
	} else {
		c.evict()
		c.cache[key] = c.hand
		c.hand.Value.(*entry).AccessCounter = c.clockCount
		c.hand.Value.(*entry).Key = key
		c.hand.Value.(*entry).Value = value
		c.hand = c.hand.Next()
	}
}

// Get look up a key in the cache. The second return value indicates if the key was found.
func (c *ClockCache) Get(key interface{}) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if e, ok := c.cache[key]; ok {
		if e.Value.(*entry).AccessCounter != c.clockCount {
			e.Value.(*entry).AccessCounter = c.clockCount
		}
		return e.Value.(*entry).Value, true
	}
	return nil, false
}

// Delete removes an entry with the given key from the cache.
func (c *ClockCache) Delete(key interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if e, ok := c.cache[key]; ok {
		delete(c.cache, key)
		e.Value.(*entry).AccessCounter = 0
	}
}

func (c *ClockCache) evict() {
	for c.hand.Value.(*entry).AccessCounter > 0 {
		c.hand.Value.(*entry).AccessCounter--
		c.hand = c.hand.Next()
	}
	if c.hand.Value.(*entry).Key != nil {
		delete(c.cache, c.hand.Value.(*entry).Key)
		if c.onEvict != nil {
			go c.onEvict(c.hand.Value.(*entry).Key, c.hand.Value.(*entry).Value)
		}
	}
}