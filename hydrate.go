package ecorder

import (
	"github.com/evecentral/eccore"
	"github.com/theatrus/crestmarket"
)

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
func NewCrestHydrator(req crestmarket.CRESTRequestor, static eccore.StaticItems) (Hydrator, error) {
	ch := &crestHydrator{crest: req, static: static}

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
