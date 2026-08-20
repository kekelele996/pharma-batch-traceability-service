package httpapi

import (
	"net/http"

	"pharma-batch-traceability-service/internal/relocation"
	"pharma-batch-traceability-service/internal/platform"
)

func (s *Server) createDispatch(w http.ResponseWriter, r *http.Request) {
	var v relocation.Dispatch
	if err := platform.DecodeJSON(r, &v); err != nil {
		platform.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	created, err := s.dispatch.Create(v)
	if err != nil {
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusCreated, created)
}

func (s *Server) listDispatch(w http.ResponseWriter, r *http.Request) {
	platform.WriteJSON(w, http.StatusOK, s.dispatch.List())
}

func (s *Server) startDispatch(w http.ResponseWriter, r *http.Request) {
	id := pathID(r, "id")
	if !dispatchTryStart(id) {
		platform.WriteError(w, http.StatusConflict, "dispatch already started")
		return
	}
	updated, err := s.dispatch.Start(id)
	if err != nil {
		dispatchRestore(id)
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusOK, updated)
}

func (s *Server) completeDispatch(w http.ResponseWriter, r *http.Request) {
	id := pathID(r, "id")
	if !dispatchRelease(id) {
		platform.WriteError(w, http.StatusConflict, "dispatch not started")
		return
	}
	updated, err := s.dispatch.Complete(id)
	if err != nil {
		dispatchRestore(id)
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusOK, updated)
}

func (s *Server) cancelDispatch(w http.ResponseWriter, r *http.Request) {
	id := pathID(r, "id")
	if !dispatchRelease(id) {
		platform.WriteError(w, http.StatusConflict, "dispatch not started")
		return
	}
	updated, err := s.dispatch.Cancel(id)
	if err != nil {
		dispatchRestore(id)
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusOK, updated)
}

func (s *Server) placeQuarantine(w http.ResponseWriter, r *http.Request) {
	var body struct {
		BatchID string `json:"batch_id"`
		Reason  string `json:"reason"`
	}
	if err := platform.DecodeJSON(r, &body); err != nil {
		platform.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	created, err := s.quarantine.Place(body.BatchID, body.Reason)
	if err != nil {
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusCreated, created)
}

func (s *Server) listQuarantine(w http.ResponseWriter, r *http.Request) {
	platform.WriteJSON(w, http.StatusOK, s.quarantine.ListByBatch(r.URL.Query().Get("batch_id")))
}

func (s *Server) releaseQuarantine(w http.ResponseWriter, r *http.Request) {
	updated, err := s.quarantine.Release(pathID(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusOK, updated)
}

func (s *Server) rejectQuarantine(w http.ResponseWriter, r *http.Request) {
	updated, err := s.quarantine.Reject(pathID(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusOK, updated)
}

func (s *Server) issueRecall(w http.ResponseWriter, r *http.Request) {
	var body struct {
		BatchID string `json:"batch_id"`
		Level   int    `json:"level"`
		Reason  string `json:"reason"`
	}
	if err := platform.DecodeJSON(r, &body); err != nil {
		platform.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	created, err := s.recall.Issue(body.BatchID, body.Level, body.Reason)
	if err != nil {
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusCreated, created)
}

func (s *Server) listRecall(w http.ResponseWriter, r *http.Request) {
	platform.WriteJSON(w, http.StatusOK, s.recall.List())
}

func (s *Server) recallList(w http.ResponseWriter, r *http.Request) {
	list, err := s.recall.GenerateList(pathID(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusOK, list)
}

func (s *Server) executeRecall(w http.ResponseWriter, r *http.Request) {
	updated, err := s.recall.Execute(pathID(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusOK, updated)
}

func (s *Server) completeRecall(w http.ResponseWriter, r *http.Request) {
	updated, err := s.recall.Complete(pathID(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusOK, updated)
}

