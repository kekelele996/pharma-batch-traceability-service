package lineage

import (
	"pharma-batch-traceability-service/internal/production"
	"pharma-batch-traceability-service/internal/serialization"
	"pharma-batch-traceability-service/internal/shipment"
	"pharma-batch-traceability-service/internal/stock"
)

type Service struct {
	batches    *production.Service
	serials    *serialization.Service
	shipments  *shipment.Service
	stock      *stock.Service
}

func NewService(b *production.Service, srl *serialization.Service, sh *shipment.Service, st *stock.Service) *Service {
	return &Service{batches: b, serials: srl, shipments: sh, stock: st}
}

// TraceByCode resolves a full serial code (GTIN+serial) to its complete
// traceability chain: production batch, movement timeline and current stock.
func (s *Service) TraceByCode(code string) (TraceResult, error) {
	srl, err := s.serials.Get(code)
	if err != nil {
		return TraceResult{}, err
	}
	res, err := s.TraceByBatch(srl.BatchID)
	if err != nil {
		return TraceResult{}, err
	}
	res.Code = code
	return res, nil
}

// TraceByBatch assembles the chain for a production batch and verifies chain
// integrity: the first movement must be production, and any released batch
// without a production node is flagged as an incomplete chain.
func (s *Service) TraceByBatch(batchID string) (TraceResult, error) {
	b, err := s.batches.Get(batchID)
	if err != nil {
		return TraceResult{}, err
	}
	res := TraceResult{
		BatchID:   b.ID,
		DrugID:    b.DrugID,
		BatchNo:   b.BatchNo,
		Status:    b.Status,
		Timeline:  s.shipments.ListByBatch(batchID),
		Positions: s.stock.ListByBatch(batchID),
		Complete:  true,
	}
	if len(res.Timeline) == 0 {
		res.Complete = false
		res.Gaps = append(res.Gaps, "no movement recorded")
		return res, nil
	}
	if res.Timeline[0].NodeType != shipment.NodeProduction {
		res.Complete = false
		res.Gaps = append(res.Gaps, "missing production origin")
	}
	for i := 1; i < len(res.Timeline); i++ {
		prev, cur := res.Timeline[i-1], res.Timeline[i]
		if prev.ToID != "" && cur.FromID != "" && prev.ToID != cur.FromID {
			res.Complete = false
			res.Gaps = append(res.Gaps, "broken handoff between "+prev.ID+" and "+cur.ID)
		}
	}
	return res, nil
}

