package inbound

// rollback reverses the stock already received for items in a failed inbound,
// so a batch that aborts the operation does not leave earlier shelves stocked.
// Items that were never received (or no longer hold enough quantity) are
// skipped rather than aborting the rollback of the rest.
func (s *Service) rollback(warehouseID string, items []InboundItem) {
	for _, it := range items {
		_, _ = s.stock.Reverse(it.BatchID, warehouseID, it.Qty)
	}
}
