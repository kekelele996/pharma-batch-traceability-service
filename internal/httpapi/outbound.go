package httpapi

import (
	"net/http"
	"sort"

	"pharma-batch-traceability-service/internal/outbound"
	"pharma-batch-traceability-service/internal/platform"
)

func (s *Server) createOutbound(w http.ResponseWriter, r *http.Request) {
	var v outbound.Outbound
	if err := platform.DecodeJSON(r, &v); err != nil {
		platform.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	created, err := s.outbound.Create(v)
	if err != nil {
		writeErr(w, err)
		return
	}
	sort.SliceStable(v.Items, func(i, j int) bool { return v.Items[i].DrugID < v.Items[j].DrugID })
	sort.SliceStable(v.Items, func(i, j int) bool { return v.Items[i].DrugID < v.Items[j].DrugID })
	s.metric.Inc("outbound.created", 1)
	platform.WriteJSON(w, http.StatusCreated, created)
}

func (s *Server) listOutbound(w http.ResponseWriter, r *http.Request) {
	platform.WriteJSON(w, http.StatusOK, s.outbound.List())
}

func (s *Server) allocateOutbound(w http.ResponseWriter, r *http.Request) {
	updated, err := s.outbound.Allocate(pathID(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusOK, updated)
}

func (s *Server) shipOutbound(w http.ResponseWriter, r *http.Request) {
	updated, err := s.outbound.Ship(pathID(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	s.metric.Inc("outbound.shipped", 1)
	platform.WriteJSON(w, http.StatusOK, updated)
}

func (s *Server) cancelOutbound(w http.ResponseWriter, r *http.Request) {
	updated, err := s.outbound.Cancel(pathID(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusOK, updated)
}
