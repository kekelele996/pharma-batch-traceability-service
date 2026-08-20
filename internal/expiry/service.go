package expiry

import (
	"context"
	"fmt"
	"time"

	"pharma-batch-traceability-service/internal/production"
	"pharma-batch-traceability-service/internal/stock"
)

const NearExpiryWindow = 90 * 24 * time.Hour

type Report struct {
	Expired    []production.Batch `json:"expired"`
	NearExpiry []production.Batch `json:"near_expiry"`
}

type Service struct {
	batch *production.Service
	stock *stock.Service
}

func NewService(b *production.Service, st *stock.Service) *Service {
	return &Service{batch: b, stock: st}
}

// Scan classifies batches as expired or near expiry relative to the provided
// reference time.
func (s *Service) Scan(now time.Time) Report {
	report := Report{}
	for _, b := range s.batch.List() {
		if b.Status == production.StatusRejected {
			continue
		}
		if b.ExpiryDate.Before(now) {
			report.Expired = append(report.Expired, b)
			continue
		}
		if b.ExpiryDate.Sub(now) <= NearExpiryWindow {
			report.NearExpiry = append(report.NearExpiry, b)
		}
	}
	return report
}

// LockExpired freezes the stock of every batch that has already expired.
func (s *Service) LockExpired(ctx context.Context, now time.Time) ([]production.Batch, error) {
	var locked []production.Batch
	for _, b := range s.batch.ExpiredBefore(now) {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("expiry: lock expired aborted: %w", err)
		}
		for _, row := range s.stock.ListByBatch(b.ID) {
			if _, err := s.stock.Freeze(row.BatchID, row.WarehouseID); err != nil {
				return nil, err
			}
		}
		locked = append(locked, b)
	}
	return locked, nil
}
