package ecorder

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/evecentral/eccore"
	"github.com/theatrus/crestmarket"
	"time"
)

var (
	// Error for no hydration source
	ErrorNoHydrationSource = errors.New("Can't hydrate a missing cache order")

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
	return nil, nil
}

type OrderCache struct {
	mc       *memcache.Client
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
	if err == memcache.ErrCacheMiss {
		// Hydrate
	} else if err != nil {
		return nil, err
	} else {
		orderEntry, err := unpackOrder(item.Value)
		if err != nil {
			return nil, err
		}
		// Entry too old?
		if orderEntry.At.Add(cacheExpires).Before(time.Now()) {
			// hydrate
		} else {
			return orderEntry.Orders, nil
		}
	}

	return nil, nil
}
