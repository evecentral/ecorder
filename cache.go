package ecorder

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/evecentral/eccore"
	"github.com/theatrus/gomemcache/memcache"
	"time"
)

var (
	// Error for no hydration source
	ErrorNoHydrationSource = errors.New("Can't hydrate a missing cache order")
	// ErrorUnfetchableType defines an unfetchable type or region
	ErrorUnfetchableItem = errors.New("Can't fetch this type or region since its not one known to me")

	// Default cache expiration
	cacheExpires = 5 * time.Minute
)

type orderEntry struct {
	Orders []eccore.MarketOrder
	At     time.Time
}

// OrderCache represents a caching order
// provider, with a memcache client or mock
// and an optional Hydrator interface for cache
// misses
type OrderCache struct {
	Mc       memcache.Client
	Hydrator Hydrator
}

// cacheKey generates a single memcache key for an order
// stored in a region and type
func cacheKey(typeid int, regionid int) string {
	return fmt.Sprintf("ecorder/liveorders/1/%d/%d", typeid, regionid)
}

func packOrder(entry orderEntry) []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(entry)
	return buf.Bytes()
}

func unpackOrder(value []byte) (orderEntry, error) {
	dec := gob.NewDecoder(bytes.NewReader(value))
	var entry orderEntry
	err := dec.Decode(&entry)
	return entry, err
}

func (c *OrderCache) hydrateSingleOrder(typeid int, regionid int) ([]eccore.MarketOrder, error) {
	if c.Hydrator == nil {
		return nil, ErrorNoHydrationSource
	}

	orders, err := c.Hydrator.OrdersForType(typeid, regionid)
	if err != nil {
		return nil, err
	}

	key := cacheKey(typeid, regionid)
	entry := orderEntry{Orders: orders, At: time.Now()}
	encodedOrders := packOrder(entry)

	c.Mc.Set(&memcache.Item{
		Key:        key,
		Value:      encodedOrders,
		Expiration: int32(cacheExpires.Seconds())})

	return orders, nil
}

func (c *OrderCache) OrdersForType(typeid int, regionid int) ([]eccore.MarketOrder, error) {
	if typeid <= 0 || regionid <= 0 {
		return nil, errors.New("Flagantrantly wrong error")
	}

	key := cacheKey(typeid, regionid)

	item, err := c.Mc.Get(key)

	if err != nil && err != memcache.ErrCacheMiss {
		return nil, err
	} else if err == memcache.ErrCacheMiss {
		// Cache didn't return, now we hydrate and return
		return c.hydrateSingleOrder(typeid, regionid)
	} else {
		orderEntry, err := unpackOrder(item.Value)
		if err != nil {
			return nil, err
		}
		// Entry too old?
		if !orderEntry.At.Add(cacheExpires).After(time.Now()) {
			return orderEntry.Orders, nil
		} else {
			return c.hydrateSingleOrder(typeid, regionid)
		}
	}
}
