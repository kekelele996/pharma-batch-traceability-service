package httpapi

import (
	"net/http"

	"pharma-batch-traceability-service/internal/platform"
)

func (s *Server) listStock(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if batchID := q.Get("batch_id"); batchID != "" {
		platform.WriteJSON(w, http.StatusOK, s.stock.ListByBatch(batchID))
		return
	}
	if warehouseID := q.Get("warehouse_id"); warehouseID != "" {
		platform.WriteJSON(w, http.StatusOK, s.stock.ListByWarehouse(warehouseID))
		return
	}
	platform.WriteJSON(w, http.StatusOK, s.stock.List())
}
