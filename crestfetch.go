package ecorder

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/evecentral/eccore"
	"github.com/theatrus/crestmarket"
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

type Hydrator interface {
	OrdersForType(typeid int, regionid int) ([]eccore.MarketOrder, error)
}

type crestHydrator struct {
	crest   crestmarket.CRESTRequestor
	types   *crestmarket.MarketTypes
	regions *crestmarket.Regions
	static  eccore.StaticItems
}

// NewCrestHydrator bootstraps an appropiate CREST
// based order hydration source.
func NewCrestHydrator(req crestmarket.CRESTRequestor) (Hydrator, error) {
	ch := &crestHydrator{crest: req}

	types, err := req.Types()
	if err != nil {
		return nil, err
	}

	regions, err := req.Regions()
	if err != nil {
		return nil, err
	}

	ch.types = types
	ch.regions = regions

	return ch, nil
}

func (ch *crestHydrator) OrdersForType(typeid int, regionid int) ([]eccore.MarketOrder, error) {
	itemType := ch.types.ById(typeid)
	region := ch.regions.ById(regionid)

	if itemType == nil || region == nil {
		return nil, ErrorUnfetchableItem
	}

	orders, err := ch.crest.BuySellMarketOrders(region, itemType)
	if err != nil {
		return nil, err
	}
	rOrders := make([]eccore.MarketOrder, len(orders.Orders))
	for i, order := range orders.Orders {
		rOrders[i] = eccore.CRESTToOrder(order, ch.static)
	}
	return rOrders, nil
}

// OrderCache represents a caching order
// provider, with a memcache client or mock
// and an optional Hydrator interface for cache
// misses
type OrderCache struct {
	mc       memcache.Client
	hydrator Hydrator
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
	if c.hydrator == nil {
		return nil, ErrorNoHydrationSource
	}

	return nil, nil
}

func (c *OrderCache) OrdersForType(typeid int, regionid int) ([]eccore.MarketOrder, error) {
	if typeid <= 0 || regionid <= 0 {
		return nil, errors.New("Flagantrantly wrong error")
	}

	key := cacheKey(typeid, regionid)

	item, err := c.mc.Get(key)

	if err != nil && err != memcache.ErrCacheMiss {
		return nil, err
	} else {
		orderEntry, err := unpackOrder(item.Value)
		if err != nil {
			return nil, err
		}
		// Entry too old?
		if orderEntry.At.Add(cacheExpires).After(time.Now()) {
			return orderEntry.Orders, nil
		}
	}

	// Cache didn't return, now we hydrate and return

	return nil, nil
}
