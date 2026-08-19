package inbound

// rollback is supposed to reverse stock already received for a failed inbound,
// but the current implementation does nothing.
func (s *Service) rollback(warehouseID string, items []InboundItem) {
}
