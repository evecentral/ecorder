package ecorder

import (
	"github.com/evecentral/sdetools"
	"github.com/evecentral/eccore"
)

type Hydrator interface {
	OrdersForType(typeid int, regionid int) ([]eccore.MarketOrder, error)
	OrdersForRegion(regionid int) ([]eccore.MarketOrder, error)
}

type esiHydrator struct {
	sde *sdetools.SDE
}

// NewCrestHydrator bootstraps an appropiate CREST
// based order hydration source.
func NewESIHydrator() (Hydrator, error) {
	return nil, nil
}

func (ch *esiHydrator) OrdersForType(typeid int, regionid int) ([]eccore.MarketOrder, error) {
	return nil, nil
}
