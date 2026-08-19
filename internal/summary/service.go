package summary

import (
	"time"

	"pharma-batch-traceability-service/internal/production"
	"pharma-batch-traceability-service/internal/serialization"
	"pharma-batch-traceability-service/internal/stock"
	"pharma-batch-traceability-service/internal/util"
)

type Summary struct {
	TotalBatches    int `json:"total_batches"`
	ReleasedBatches int `json:"released_batches"`
	RejectedBatches int `json:"rejected_batches"`
	ExpiredBatches  int `json:"expired_batches"`
	TotalStock      int `json:"total_stock"`
	AvailableStock  int `json:"available_stock"`
	LockedStock     int `json:"locked_stock"`
	TotalSerials    int `json:"total_serials"`
	ActiveSerials   int `json:"active_serials"`
	SoldSerials     int `json:"sold_serials"`
}

type Service struct {
	stock   *stock.Service
	batches *production.Service
	serials *serialization.Service
}

func NewService(st *stock.Service, b *production.Service, srl *serialization.Service) *Service {
	return &Service{stock: st, batches: b, serials: srl}
}

func (s *Service) Summary(now time.Time) Summary {
	var sum Summary
	batches := s.batches.List()
	released := util.Filter(batches, func(b production.Batch) bool { return b.Status == production.StatusReleased })
	rejected := util.Filter(released, func(b production.Batch) bool { return b.Status == production.StatusRejected })
	expired := util.Filter(batches, func(b production.Batch) bool { return b.ExpiryDate.Before(now) })
	sum.TotalBatches = len(batches)
	sum.ReleasedBatches = len(released)
	sum.RejectedBatches = len(rejected)
	sum.ExpiredBatches = len(expired)
	for _, row := range s.stock.List() {
		sum.TotalStock += row.Quantity
		sum.AvailableStock += row.Available()
		sum.LockedStock += row.Locked
	}
	for _, srl := range s.serials.ListByBatch("") {
		sum.TotalSerials++
		switch srl.Status {
		case serialization.StatusActive:
			sum.ActiveSerials++
		case serialization.StatusSold:
			sum.SoldSerials++
		}
	}
	return sum
}
