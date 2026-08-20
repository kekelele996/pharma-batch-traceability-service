package httpapi

import (
	"net/http"

	"pharma-batch-traceability-service/internal/production"
	"pharma-batch-traceability-service/internal/platform"
)

func (s *Server) createBatch(w http.ResponseWriter, r *http.Request) {
	var b production.Batch
	if err := platform.DecodeJSON(r, &b); err != nil {
		platform.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	created, err := s.batches.Create(b)
	if err != nil {
		writeErr(w, err)
		return
	}
	s.metric.Inc("batches.created", 1)
	platform.WriteJSON(w, http.StatusCreated, created)
}

func (s *Server) listBatches(w http.ResponseWriter, r *http.Request) {
	if drugID := r.URL.Query().Get("drug_id"); drugID != "" {
		platform.WriteJSON(w, http.StatusOK, s.batches.ListByDrug(drugID))
		return
	}
	if status := r.URL.Query().Get("status"); status != "" {
		platform.WriteJSON(w, http.StatusOK, s.batches.ListByStatus(status))
		return
	}
	platform.WriteJSON(w, http.StatusOK, s.batches.List())
}

func (s *Server) getBatch(w http.ResponseWriter, r *http.Request) {
	b, err := s.batches.Get(pathID(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusOK, b)
}

func (s *Server) releaseBatch(w http.ResponseWriter, r *http.Request) {
	id := pathID(r, "id")
	setOpFlag(releaseOps, id)
	b, err := s.batches.Release(id)
	if err != nil {
		clearOpFlag(releaseOps, id)
		writeErr(w, err)
		return
	}
	s.audit.Append("api", "production.release", b.ID, b.BatchNo)
	platform.WriteJSON(w, http.StatusOK, b)
}

func (s *Server) rejectBatch(w http.ResponseWriter, r *http.Request) {
	b, err := s.batches.Reject(pathID(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusOK, b)
}

func (s *Server) quarantineBatch(w http.ResponseWriter, r *http.Request) {
	b, err := s.batches.Quarantine(pathID(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusOK, b)
}
