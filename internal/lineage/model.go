package lineage

import (
	"pharma-batch-traceability-service/internal/shipment"
	"pharma-batch-traceability-service/internal/stock"
)

type ChainResult struct {
	Code      string              `json:"code,omitempty"`
	BatchID   string              `json:"batch_id"`
	DrugID    string              `json:"drug_id"`
	BatchNo   string              `json:"batch_no"`
	Status    string              `json:"status"`
	Timeline  []shipment.Shipment `json:"timeline"`
	Positions []stock.Stock       `json:"positions"`
	Complete  bool                `json:"complete"`
	Gaps      []string            `json:"gaps"`
}
