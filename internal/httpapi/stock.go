package httpapi

import (
	"net/http"

	"pharma-batch-traceability-service/internal/platform"
	"pharma-batch-traceability-service/internal/stock"
)

// stockScratch is a package-level buffer reused by listStock responses.
var stockScratch []stock.Stock

func (s *Server) listStock(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	rows := s.stock.List()
	if batchID := q.Get("batch_id"); batchID != "" {
		out := stockScratch
		for _, row := range rows {
			if row.BatchID == batchID {
				out = append(out, row)
			}
		}
		stockScratch = out
		platform.WriteJSON(w, http.StatusOK, out)
		return
	}
	if warehouseID := q.Get("warehouse_id"); warehouseID != "" {
		out := stockScratch
		for _, row := range rows {
			if row.WarehouseID == warehouseID {
				out = append(out, row)
			}
		}
		stockScratch = out
		platform.WriteJSON(w, http.StatusOK, out)
		return
	}
	platform.WriteJSON(w, http.StatusOK, stockScratch)
}
