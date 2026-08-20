package httpapi

import (
	"net/http"

	"pharma-batch-traceability-service/internal/platform"
	"pharma-batch-traceability-service/internal/quality"
	"time"
)

func (s *Server) createQuality(w http.ResponseWriter, r *http.Request) {
	var v quality.Record
	if err := platform.DecodeJSON(r, &v); err != nil {
		platform.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	created, err := s.quality.Create(v)
	if err != nil {
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusCreated, created)
}

func (s *Server) listQuality(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if id := q.Get("batch_id"); id != "" {
		platform.WriteJSON(w, http.StatusOK, s.quality.ListByBatch(id))
		return
	}
	if id := q.Get("inbound_id"); id != "" {
		platform.WriteJSON(w, http.StatusOK, s.quality.ListByInbound(id))
		return
	}
	platform.WriteJSON(w, http.StatusOK, s.quality.ListByBatch(""))
}

func (s *Server) decideQuality(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Result    string `json:"result"`
		Inspector string `json:"inspector"`
		Note      string `json:"note"`
	}
	if err := platform.DecodeJSON(r, &body); err != nil {
		platform.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	updated, err := s.quality.Decide(pathID(r, "id"), body.Result, body.Inspector, body.Note)
	if err != nil {
		platform.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	platform.WriteJSON(w, http.StatusOK, updated)
}

func (s *Server) recordColdchain(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ShipmentID string  `json:"shipment_id"`
		DeviceID   string  `json:"device_id"`
		Storage    string  `json:"storage"`
		TempC      float64 `json:"temp_c"`
		Humidity   float64 `json:"humidity"`
	}
	if err := platform.DecodeJSON(r, &body); err != nil {
		platform.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	reading, err := s.coldchain.Record(body.ShipmentID, body.DeviceID, body.Storage, body.TempC, body.Humidity, time.Time{})
	if err != nil {
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusCreated, reading)
}

func (s *Server) listColdchain(w http.ResponseWriter, r *http.Request) {
	platform.WriteJSON(w, http.StatusOK, s.coldchain.ListByShipment(r.URL.Query().Get("shipment_id")))
}

func (s *Server) enqueueNotification(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Channel   string `json:"channel"`
		Recipient string `json:"recipient"`
		Subject   string `json:"subject"`
		Body      string `json:"body"`
	}
	if err := platform.DecodeJSON(r, &body); err != nil {
		platform.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	n := s.notification.Enqueue(body.Channel, body.Recipient, body.Subject, body.Body)
	platform.WriteJSON(w, http.StatusCreated, n)
}

func (s *Server) sendNotifications(w http.ResponseWriter, r *http.Request) {
	sent := s.notification.SendPending()
	platform.WriteJSON(w, http.StatusOK, sent)
}

func (s *Server) listNotifications(w http.ResponseWriter, r *http.Request) {
	platform.WriteJSON(w, http.StatusOK, s.notification.List(100))
}

func (s *Server) generateLabels(w http.ResponseWriter, r *http.Request) {
	var body struct {
		BatchID string `json:"batch_id"`
		Qty     int    `json:"qty"`
	}
	if err := platform.DecodeJSON(r, &body); err != nil {
		platform.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	labels, err := s.label.Generate(body.BatchID, body.Qty)
	if err != nil {
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusCreated, labels)
}

func (s *Server) printLabel(w http.ResponseWriter, r *http.Request) {
	updated, err := s.label.Print(pathID(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	platform.WriteJSON(w, http.StatusOK, updated)
}

func (s *Server) listLabels(w http.ResponseWriter, r *http.Request) {
	platform.WriteJSON(w, http.StatusOK, s.label.ListByBatch(r.URL.Query().Get("batch_id")))
}

func (s *Server) reportSummary(w http.ResponseWriter, r *http.Request) {
	platform.WriteJSON(w, http.StatusOK, s.report.Summary(time.Now().UTC()))
}

