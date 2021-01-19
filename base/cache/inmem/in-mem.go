// Package inmem implements in-memory cache for storing data
package inmem

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

type (
	item struct {
		expired bool
		object  interface{}
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
		onExpired  func(string, interface{})
		onEvicted  func(string, interface{})
		onAdd      func(string, interface{})
		onDelete   func(string, interface{})
		onSet      func(string, interface{})
		janitor    *janitor
	}

	keyval struct {
		key   string
		value interface{}
	}

	Option func(*cache)
)

func (i *item) Value() interface{} { return i.object }
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
func (c *cache) delete(k string) (interface{}, bool) {
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
func (c *cache) Set(k string, val interface{}) {
	c.mutex.Lock()
	c.set(k, val)
	c.mutex.Unlock()
	if c.onSet != nil {
		c.onSet(k, val)
	}
}

// Add an item to the cache only if an item doesn't exist for the given key
// or if the existing item has expired. Returns error otherwise
func (c *cache) Add(k string, val interface{}) error {
	c.mutex.Lock()
	_, found := c.get(k)
	if found {
		c.mutex.Unlock()
		return fmt.Errorf("Item %s already exists", k)
	}

	c.set(k, val)
	c.mutex.Unlock()
	if c.onAdd != nil {
		c.onAdd(k, val)
	}
	return nil
}

// Replace item if it exists
func (c *cache) Replace(k string, val interface{}) error {
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

func (c *cache) set(k string, val interface{}) {
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
	k string,
	val interface{},
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
func (c *cache) get(k string) (interface{}, bool) {
	item, found := c.items[k]

	if !found || item.expired {
		return nil, false
	}

	return item.object, true
}

func (c *cache) Get(k string) (interface{}, bool) {
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

func (c *cache) Delete(key string) {
	c.mutex.Lock()
	v, evicted := c.delete(key)
	c.mutex.Unlock()

	if evicted {
		c.onEvicted(key, v)
	}

	if c.onDelete != nil {
		c.onDelete(key, v)
	}
}

func (c *cache) print() {
	for k, v := range c.items {
		fmt.Println("key:", k, "expired:", v.expired, "expiry:", v.expires, "evict:", v.evicts)
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

//Load initializes a cache
func (c *cache) Load(values map[string]interface{}) {
	for k, v := range values {
		c.Set(k, v)
	}
}

//Reload replaces internal map with a new one
func (c *cache) Reload(values map[string]interface{}) {
	var removed []string
	for k := range c.items {
		if _, ok := values[k]; !ok {
			//key not present in newer map
			removed = append(removed, k)
		}
	}

	for _, k := range removed {
		c.Delete(k)
	}

	var added []string
	for k := range values {
		if _, ok := c.items[k]; !ok {
			//key not present in current map
			added = append(added, k)
		}
	}

	for _, k := range added {
		c.Add(k, values[k])
	}

	for k, v := range values {
		if oldvalue, ok := c.Get(k); ok {
			if v != oldvalue {
				c.Set(k, v)
			}
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

func (c *cache) OnExpired(fn func(string, interface{})) {
	c.mutex.Lock()
	c.onExpired = fn
	c.mutex.Unlock()
}

func (c *cache) OnEvicted(fn func(string, interface{})) {
	c.mutex.Lock()
	c.onEvicted = fn
	c.mutex.Unlock()
}

func (c *cache) OnAdd(fn func(string, interface{})) {
	c.mutex.Lock()
	c.onAdd = fn
	c.mutex.Unlock()
}

func (c *cache) OnDelete(fn func(string, interface{})) {
	c.mutex.Lock()
	c.onDelete = fn
	c.mutex.Unlock()
}

func (c *cache) OnSet(fn func(string, interface{})) {
	c.mutex.Lock()
	c.onSet = fn
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

//WithOnEvictCallback allows configuring a fn to be called on eviction
func WithOnEvictCallback(fn func(k string, val interface{})) Option {
	return func(c *cache) {
		c.onEvicted = fn
	}
}

//WithOnExpiredCallback allows configuring a fn to be called on expiry
func WithOnExpiredCallback(fn func(k string, val interface{})) Option {
	return func(c *cache) {
		c.onExpired = fn
	}
}

//WithOnAddCallback allows configuring a fn to be called on add
func WithOnAddCallback(fn func(k string, val interface{})) Option {
	return func(c *cache) {
		c.onAdd = fn
	}
}

//WithOnSetCallback allows configuring a fn to be called on set
func WithOnSetCallback(fn func(k string, val interface{})) Option {
	return func(c *cache) {
		c.onSet = fn
	}
}

//WithOnDeleteCallback allows configuring a fn to be called on delete
func WithOnDeleteCallback(fn func(k string, val interface{})) Option {
	return func(c *cache) {
		c.onDelete = fn
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
