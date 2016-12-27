package ecorder

import (
	"errors"
	"fmt"

	"log"
	"time"

	"github.com/evecentral/eccore"

	"gopkg.in/redis.v5"
	"gopkg.in/vmihailenco/msgpack.v2"
)

var (
	// Error for no hydration source
	ErrorNoHydrationSource = errors.New("Can't hydrate a missing cache order")
	// ErrorUnfetchableType defines an unfetchable type or region
	ErrorUnfetchableItem = errors.New("Can't fetch this type or region since its not one known to me")

	// Default cache expiration
	cacheExpires = 5 * time.Minute

	// Memcache expiration for the key
	cacheMcExpires = 6 * time.Hour
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
	redis    redis.Client
	Hydrator Hydrator
}

// cacheKey generates a single memcache key for an order
// stored in a region and type
func cacheKey(typeid int, regionid int) string {
	return fmt.Sprintf("ecorder/liveorders/1/%d/%d", typeid, regionid)
}

func packOrder(entry orderEntry) []byte {
	b, _ := msgpack.Marshal(&entry)
	return b
}

func unpackOrder(value []byte) (oe *orderEntry, err error) {
	err = msgpack.Unmarshal(value, &oe)
	return
}

func (c *OrderCache) hydrateSingleOrder(typeid int, regionid int) ([]eccore.MarketOrder, error) {
	if c.Hydrator == nil {
		return nil, ErrorNoHydrationSource
	}

	orders, err := c.Hydrator.OrdersForType(typeid, regionid)
	if err != nil {
		log.Printf("hydration error: %s", err)
		return nil, err
	}

	key := cacheKey(typeid, regionid)
	entry := orderEntry{Orders: orders, At: time.Now()}
	encodedOrders := packOrder(entry)

	err = c.redis.Set(key, encodedOrders, cacheMcExpires).Err()

	return orders, nil
}

func (c *OrderCache) OrdersForType(typeid int, regionid int) ([]eccore.MarketOrder, error) {
	if typeid <= 0 || regionid <= 0 {
		return nil, errors.New("Flagantrantly wrong error")
	}

	key := cacheKey(typeid, regionid)

	item, err := c.redis.Get(key).Bytes()

	if err != nil && err != redis.Nil {
		return nil, err
	} else if err == redis.Nil {
		// Cache didn't return, now we hydrate and return
		return c.hydrateSingleOrder(typeid, regionid)
	} else {
		orderEntry, err := unpackOrder(item)
		if err != nil {
			return nil, err
		}

		// Entry too old?
		if orderEntry.At.Add(cacheExpires).After(time.Now()) {
			return orderEntry.Orders, nil
		} else {
			// We will hydrate an order asynchronously, but
			// also return the current cached value.
			go c.hydrateSingleOrder(typeid, regionid)
			return orderEntry.Orders, nil
		}
	}
}
