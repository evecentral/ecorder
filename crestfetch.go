package ecorder

import (
	"errors"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/theatrus/crestmarket"
	"log"
)

type crestCache struct {
	mc    *memcache.Client
	crest crestmarket.CRESTRequestor
}

func (c *crestCache) Orders(typeid int, regionid int) (MarketOrder, error) {
	if typeid <= 0 || regionid <= 0 {
		return nil, errors.New("Flagantrantly wrong error")
	}
	return nil, nil
}
