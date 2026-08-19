package inbound

// rollback reverses stock already received for an inbound that failed partway.
func (s *Service) rollback(warehouseID string, items []InboundItem) {
	for _, it := range items {
		_, _ = s.stock.Deduct(it.BatchID, warehouseID, it.Qty)
	}
}
