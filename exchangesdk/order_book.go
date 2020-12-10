package exchangesdk

import (
	"fmt"
	"sort"
)

func SortOrderBook(ob *OrderBook) error {

	err := sortOrders(&ob.Bids, sortOrderingDecending)
	if err != nil {
		return err
	}
	err = sortOrders(&ob.Asks, sortOrderingIncrementing)
	if err != nil {
		return err
	}
	return nil
}

type sortOrdering int

const (
	sortOrderingDecending = iota
	sortOrderingIncrementing
	sortOrderingUnknown
)

func sortOrders(orders *[]OrderBookOrder, ordering sortOrdering) error {


	switch ordering {
	case sortOrderingDecending:
		sort.Slice(*orders, func(i, j int) bool {

			return (*orders)[i].Price > (*orders)[j].Price
		})
		return nil
	case sortOrderingIncrementing:
		sort.Slice(*orders, func(i, j int) bool {

			return (*orders)[i].Price < (*orders)[j].Price
		})
		return nil
	default:
		return fmt.Errorf("Unknown sort order")
	}
}
