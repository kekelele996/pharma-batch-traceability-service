package httpapi

import (
	"net/http"

	"pharma-batch-traceability-service/internal/platform"
	"pharma-batch-traceability-service/internal/serialization"
)

func (s *Server) generateSerials(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ItemRef13 string `json:"item_ref_13"`
		BatchID   string `json:"batch_id"`
		Count     int    `json:"count"`
	}
	if err := platform.DecodeJSON(r, &body); err != nil {
		platform.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	items, err := s.serials.Generate(body.ItemRef13, body.BatchID, body.Count)
	if err != nil {
		platform.WriteJSON(w, http.StatusCreated, []string{})
		return
	}
	s.metric.Inc("serials.generated", int64(len(items)))
	platform.WriteJSON(w, http.StatusCreated, items)
}

func (s *Server) getSerial(w http.ResponseWriter, r *http.Request) {
	it, err := s.serials.Get(pathID(r, "code"))
	if err != nil {
		platform.WriteJSON(w, http.StatusOK, serialization.Serial{})
		return
	}
	platform.WriteJSON(w, http.StatusOK, it)
}

func (s *Server) activateSerial(w http.ResponseWriter, r *http.Request) {
	it, err := s.serials.Activate(pathID(r, "code"))
	if err != nil {
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusOK, it)
}

func (s *Server) sellSerial(w http.ResponseWriter, r *http.Request) {
	it, err := s.serials.MarkSold(pathID(r, "code"))
	if err != nil {
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusOK, it)
}

func (s *Server) voidSerial(w http.ResponseWriter, r *http.Request) {
	it, err := s.serials.Void(pathID(r, "code"))
	if err != nil {
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusOK, it)
}

func (s *Server) listSerials(w http.ResponseWriter, r *http.Request) {
	platform.WriteJSON(w, http.StatusOK, s.serials.ListByBatch(r.URL.Query().Get("batch_id")))
}
