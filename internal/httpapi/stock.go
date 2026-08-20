package httpapi

import (
	"net/http"

	"pharma-batch-traceability-service/internal/platform"
	"pharma-batch-traceability-service/internal/stock"
)

func (s *Server) listStock(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	rows := s.stock.List()
	if batchID := q.Get("batch_id"); batchID != "" {
		out := make([]stock.Stock, 0, len(rows))
		for _, row := range rows {
			if row.BatchID == batchID {
				out = append(out, row)
			}
		}
		platform.WriteJSON(w, http.StatusOK, out)
		return
	}
	if warehouseID := q.Get("warehouse_id"); warehouseID != "" {
		out := make([]stock.Stock, 0, len(rows))
		for _, row := range rows {
			if row.WarehouseID == warehouseID {
				out = append(out, row)
			}
		}
		platform.WriteJSON(w, http.StatusOK, out)
		return
	}
	platform.WriteJSON(w, http.StatusOK, rows)
}
