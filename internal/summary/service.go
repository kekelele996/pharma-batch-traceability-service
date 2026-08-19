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
	allBatches := s.batches.List()
	sum.TotalBatches = len(allBatches)
	sum.ReleasedBatches = len(util.Filter(allBatches, func(b production.Batch) bool { return b.Status == production.StatusReleased }))
	sum.RejectedBatches = len(util.Filter(allBatches, func(b production.Batch) bool { return b.Status == production.StatusRejected }))
	sum.ExpiredBatches = len(util.Filter(allBatches, func(b production.Batch) bool { return b.ExpiryDate.Before(now) }))
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
