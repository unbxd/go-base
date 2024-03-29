// Package inmem implements in-memory cache for storing data
package inmem

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

type (
	item struct {
		expired bool
		object  []byte
		expires int64
		evicts  int64
	}

	janitor struct {
		expireDuration time.Duration
		purgeDuration  time.Duration
		stop           chan bool
	}

	cache struct {
		expiration time.Duration
		eviction   time.Duration
		items      map[string]*item
		mutex      sync.RWMutex
		onExpired  func(string, []byte)
		onEvicted  func(string, []byte)
		janitor    *janitor
	}

	keyval struct {
		key   string
		value []byte
	}

	Option func(*cache)
)

func (i *item) Value() []byte      { return i.object }
func (i *item) Expired() bool      { return i.expired }
func (i *item) Expires() time.Time { return time.Unix(0, i.expires) }
func (i *item) Evicts() time.Time  { return time.Unix(0, i.evicts) }

func (j *janitor) Run(c *cache) {
	exticker := time.NewTicker(j.expireDuration)
	puticker := time.NewTicker(j.purgeDuration)

	for {
		select {
		case <-exticker.C:
			c.MarkExpired()
		case <-puticker.C:
			c.Purge()
		case <-j.stop:
			exticker.Stop()
			puticker.Stop()
			return
		}
	}
}

func (c *cache) Flush() {
	c.mutex.Lock()
	c.items = make(map[string]*item)
	c.mutex.Unlock()
}

// Returns the object value stored and if it is found
// This method is not thread safe
func (c *cache) delete(k string) ([]byte, bool) {
	if c.onEvicted != nil {
		if v, found := c.items[k]; found {
			delete(c.items, k)
			return v.object, true
		}
	}

	delete(c.items, k)
	return nil, false
}

// Adds the item to cache replacing existing one
func (c *cache) Set(_ context.Context, k string, val []byte) {
	c.mutex.Lock()
	c.set(k, val)
	// c.print()
	c.mutex.Unlock()
}

// Add an item to the cache only if an item doesn't exist for the given key
// or if the existing item has expired. Returns error otherwise
func (c *cache) Add(_ context.Context, k string, val []byte) error {
	c.mutex.Lock()
	_, found := c.get(k)
	if found {
		c.mutex.Unlock()
		return fmt.Errorf("Item %s already exists", k)
	}

	c.set(k, val)
	c.mutex.Unlock()
	return nil
}

// Replace item if it exists
func (c *cache) Replace(_ context.Context, k string, val []byte) error {
	c.mutex.Lock()
	_, found := c.get(k)
	if !found {
		c.mutex.Unlock()
		return fmt.Errorf("Item %s doesn't exist", k)
	}

	c.set(k, val)
	c.mutex.Unlock()
	return nil
}

func (c *cache) set(k string, val []byte) {
	expires := time.Now().Add(c.expiration)
	evicts := expires.Add(c.eviction)
	c.items[k] = &item{
		object:  val,
		expired: false,
		expires: expires.UnixNano(),
		evicts:  evicts.UnixNano(),
	}
}

func (c *cache) SetWithDuration(
	_ context.Context,
	k string,
	val []byte,
	expiration time.Duration,
) {
	expires := time.Now().Add(expiration)
	evicts := expires.Add(c.eviction)

	c.mutex.Lock()
	c.items[k] = &item{
		object:  val,
		expired: false,
		expires: expires.UnixNano(),
		evicts:  evicts.UnixNano(),
	}
	c.mutex.Unlock()
}

// get retrieves the item from cache, but is not thread safe
func (c *cache) get(k string) ([]byte, bool) {
	item, found := c.items[k]

	if !found || item.expired {
		return nil, false
	}

	return item.object, true
}

func (c *cache) Get(_ context.Context, k string) ([]byte, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	val, found := c.get(k)
	if !found {
		return nil, false
	}
	//c.print()
	return val, true
}

func (c *cache) GetItem(k string) (*item, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	item, found := c.items[k]
	return item, found
}

func (c *cache) Delete(_ context.Context, key string) {
	c.mutex.Lock()
	v, evicted := c.delete(key)
	c.mutex.Unlock()

	if evicted {
		c.onEvicted(key, v)
	}
}

func (c *cache) MarkExpired() {
	var expiredItems []keyval

	onExpired := (c.onExpired != nil)
	now := time.Now().UnixNano()

	c.mutex.Lock()
	for k, v := range c.items {
		if now > v.expires {
			v.expired = true
			if onExpired {
				expiredItems = append(
					expiredItems, keyval{k, v.object},
				)

			}
		}
	}
	// c.print()
	c.mutex.Unlock()

	if onExpired {
		for _, ei := range expiredItems {
			c.onExpired(ei.key, ei.value)
		}
	}
}

func (c *cache) Purge() {
	var evictedItems []keyval

	onEvicted := (c.onEvicted != nil)
	now := time.Now().UnixNano()

	c.mutex.Lock()
	for k, v := range c.items {
		if v.expired && now > v.evicts {
			val, evicted := c.delete(k)
			if evicted && onEvicted {
				evictedItems = append(
					evictedItems, keyval{k, val},
				)
			}
		}
	}
	c.mutex.Unlock()
	if onEvicted {
		for _, v := range evictedItems {
			c.onEvicted(v.key, v.value)
		}
	}
}

func (c *cache) ExpiredItems() map[string]*item {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	m := make(map[string]*item)
	for k, v := range c.items {
		if v.expired {
			m[k] = v
		}
	}
	return m
}

// Item Returns items which aren't expired
func (c *cache) Items() map[string]*item {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	m := make(map[string]*item)
	for k, v := range c.items {
		if !v.expired {
			m[k] = v
		}
	}
	return m
}

func (c *cache) OnExpired(fn func(string, []byte)) {
	c.mutex.Lock()
	c.onExpired = fn
	c.mutex.Unlock()
}

func (c *cache) OnEvicted(fn func(string, []byte)) {
	c.mutex.Lock()
	c.onEvicted = fn
	c.mutex.Unlock()
}

func newCache(
	ex time.Duration,
	ev time.Duration,
	m map[string]*item,
) *cache {
	return &cache{
		expiration: ex,
		eviction:   ev,
		items:      m,
	}
}

// Cache is sharable object which encapsulates cache
type Cache struct{ *cache }

func runJanitor(
	c *cache,
	ex time.Duration,
	ev time.Duration,
) {
	j := &janitor{
		expireDuration: ex,
		purgeDuration:  ev,
		stop:           make(chan bool),
	}

	c.janitor = j
	go j.Run(c)
}

func finalizer(c *Cache) {
	c.janitor.stop <- true
}

func newCacheWithJanitor(
	ex time.Duration,
	ev time.Duration,

	exticker time.Duration,
	evticker time.Duration,

	m map[string]*item,
) *Cache {
	c := newCache(ex, ev, m)

	C := &Cache{c}

	runJanitor(c, exticker, evticker)
	runtime.SetFinalizer(C, finalizer)

	return C
}

var (
	defaultExpiryTicker = time.Duration(10) * time.Second
	defaultEvictTicker  = time.Duration(5) * time.Minute
)

func WithOnEvictCallback(fn func(k string, val []byte)) Option {
	return func(c *cache) {
		c.onEvicted = fn
	}
}

func WithOnExpiredCallback(fn func(k string, val []byte)) Option {
	return func(c *cache) {
		c.onExpired = fn
	}
}

// New returns a new cache object
func New(
	expires time.Duration,
	evicts time.Duration,
	opts ...Option,
) *Cache {
	items := make(map[string]*item)
	c := newCacheWithJanitor(
		expires,
		evicts,
		defaultExpiryTicker,
		defaultEvictTicker,
		items,
	)

	for _, o := range opts {
		o(c.cache)
	}

	return c
}
