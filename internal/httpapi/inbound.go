package httpapi

import (
	"net/http"

	"pharma-batch-traceability-service/internal/inbound"
	"pharma-batch-traceability-service/internal/platform"
)

// inboundScratch is a package-level buffer reused by listInbound responses.
var inboundScratch []inbound.Inbound

func (s *Server) createInbound(w http.ResponseWriter, r *http.Request) {
	var v inbound.Inbound
	if err := platform.DecodeJSON(r, &v); err != nil {
		platform.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	created, err := s.inbound.Create(v)
	if err != nil {
		writeErr(w, err)
		return
	}
	s.metric.Inc("inbound.created", 1)
	platform.WriteJSON(w, http.StatusCreated, created)
}

func (s *Server) listInbound(w http.ResponseWriter, r *http.Request) {
	rows := s.inbound.List()
	if status := r.URL.Query().Get("status"); status != "" {
		out := inboundScratch
		for _, v := range rows {
			if v.Status == status {
				out = append(out, v)
			}
		}
		inboundScratch = out
		platform.WriteJSON(w, http.StatusOK, out)
		return
	}
	platform.WriteJSON(w, http.StatusOK, inboundScratch)
}

func (s *Server) acceptInbound(w http.ResponseWriter, r *http.Request) {
	var body struct {
		QCResult string `json:"qc_result"`
	}
	_ = platform.DecodeJSON(r, &body)
	updated, err := s.inbound.Accept(pathID(r, "id"), body.QCResult)
	if err != nil {
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusOK, updated)
}

func (s *Server) putawayInbound(w http.ResponseWriter, r *http.Request) {
	updated, err := s.inbound.Putaway(pathID(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	s.metric.Inc("inbound.putaway", 1)
	platform.WriteJSON(w, http.StatusOK, updated)
}
