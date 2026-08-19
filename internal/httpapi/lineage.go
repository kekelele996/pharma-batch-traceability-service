package httpapi

import (
	"net/http"
	"time"

	"pharma-batch-traceability-service/internal/platform"
)

func traceStatus(err error) int {
	switch {
	case platform.IsNotFound(err):
		return http.StatusNotFound
	case platform.IsValidation(err):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func (s *Server) traceByCode(w http.ResponseWriter, r *http.Request) {
	res, err := s.trace.ResolveByCode(pathID(r, "code"))
	if err != nil {
		platform.WriteError(w, traceStatus(err), err.Error())
		return
	}
	platform.WriteJSON(w, http.StatusOK, res)
}

func (s *Server) traceByBatch(w http.ResponseWriter, r *http.Request) {
	res, err := s.trace.ResolveByBatch(pathID(r, "batchID"))
	if err != nil {
		platform.WriteError(w, traceStatus(err), err.Error())
		return
	}
	platform.WriteJSON(w, http.StatusOK, res)
}

func (s *Server) scanExpiry(w http.ResponseWriter, r *http.Request) {
	report := s.expiry.Scan(time.Now().UTC())
	platform.WriteJSON(w, http.StatusOK, report)
}

func (s *Server) lockExpired(w http.ResponseWriter, r *http.Request) {
	locked, err := s.expiry.LockExpired(r.Context(), time.Now().UTC())
	if err != nil {
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusOK, map[string]any{"locked": len(locked)})
}

func (s *Server) listAudit(w http.ResponseWriter, r *http.Request) {
	if target := r.URL.Query().Get("target"); target != "" {
		platform.WriteJSON(w, http.StatusOK, s.audit.ListByTarget(target))
		return
	}
	platform.WriteJSON(w, http.StatusOK, s.audit.List(100))
}
