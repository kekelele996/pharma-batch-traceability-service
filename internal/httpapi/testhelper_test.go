package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pharma-batch-traceability-service/internal/audit"
	"pharma-batch-traceability-service/internal/production"
	"pharma-batch-traceability-service/internal/coldchain"
	"pharma-batch-traceability-service/internal/customer"
	"pharma-batch-traceability-service/internal/relocation"
	"pharma-batch-traceability-service/internal/drug"
	"pharma-batch-traceability-service/internal/expiry"
	"pharma-batch-traceability-service/internal/inbound"
	"pharma-batch-traceability-service/internal/label"
	"pharma-batch-traceability-service/internal/lineage"
	"pharma-batch-traceability-service/internal/manufacturer"
	"pharma-batch-traceability-service/internal/metric"
	"pharma-batch-traceability-service/internal/notification"
	"pharma-batch-traceability-service/internal/outbound"
	"pharma-batch-traceability-service/internal/platform"
	"pharma-batch-traceability-service/internal/quality"
	"pharma-batch-traceability-service/internal/isolation"
	"pharma-batch-traceability-service/internal/recall"
	"pharma-batch-traceability-service/internal/summary"
	"pharma-batch-traceability-service/internal/serialization"
	"pharma-batch-traceability-service/internal/shipment"
	"pharma-batch-traceability-service/internal/stock"
	"pharma-batch-traceability-service/internal/supplier"
	"pharma-batch-traceability-service/internal/warehouse"
)

func newTestServer() *Server {
	clock := platform.NewClock()
	drugs := drug.NewService(clock)
	manufacturers := manufacturer.NewService()
	warehouses := warehouse.NewService()
	suppliers := supplier.NewService()
	customers := customer.NewService()
	batches := production.NewService(clock)
	serials := serialization.NewService(clock)
	stocks := stock.NewService(clock)
	shipments := shipment.NewService(clock)
	tracer := lineage.NewService(batches, serials, shipments, stocks)
	inbounds := inbound.NewService(clock, batches, stocks)
	outbounds := outbound.NewService(clock, batches, stocks, shipments)
	dispatches := relocation.NewService(clock, stocks, shipments)
	quarantines := isolation.NewService(clock, batches, stocks)
	recalls := recall.NewService(clock, stocks)
	expiries := expiry.NewService(batches, stocks)
	audits := audit.NewService(clock)
	metrics := metric.NewService()
	qualitySvc := quality.NewService(clock)
	coldchainSvc := coldchain.NewService(clock)
	notificationSvc := notification.NewService(clock)
	labelSvc := label.NewService(clock, batches)
	reportSvc := summary.NewService(stocks, batches, serials)
	return NewServer(drugs, manufacturers, warehouses, suppliers, customers,
		batches, serials, stocks, shipments, tracer, inbounds, outbounds,
		dispatches, quarantines, recalls, expiries, audits, metrics,
		qualitySvc, coldchainSvc, notificationSvc, labelSvc, reportSvc)
}

func postJSON(t *testing.T, srv *Server, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)
	return rec
}

func getPath(t *testing.T, srv *Server, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)
	return rec
}

// seedTwoStockRows creates two released batches and puts both into wh-1.
func seedTwoStockRows(t *testing.T, srv *Server) {
	t.Helper()
	ids := []string{}
	for i, no := range []string{"BTA", "BTB"} {
		body := fmt.Sprintf(`{"drug_id":"d1","batch_no":"%s","production_date":"2026-07-01T00:00:00Z","expiry_date":"2027-07-01T00:00:00Z","quantity":100}`, no)
		rec := postJSON(t, srv, "/api/v1/batches", body)
		if rec.Code != http.StatusCreated {
			t.Fatalf("create batch %d: %d: %s", i, rec.Code, rec.Body.String())
		}
		var b production.Batch
		if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, b.ID)
		if rec := postJSON(t, srv, "/api/v1/batches/"+b.ID+"/release", ""); rec.Code != http.StatusOK {
			t.Fatalf("release batch: %d: %s", rec.Code, rec.Body.String())
		}
	}
	inBody := fmt.Sprintf(`{"supplier_id":"s1","warehouse_id":"wh-1","items":[{"drug_id":"d1","batch_id":"%s","qty":10},{"drug_id":"d1","batch_id":"%s","qty":20}]}`, ids[0], ids[1])
	rec := postJSON(t, srv, "/api/v1/inbound", inBody)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create inbound: %d: %s", rec.Code, rec.Body.String())
	}
	var in inbound.Inbound
	if err := json.Unmarshal(rec.Body.Bytes(), &in); err != nil {
		t.Fatal(err)
	}
	if rec := postJSON(t, srv, "/api/v1/inbound/"+in.ID+"/accept?qc_result=pass", ""); rec.Code != http.StatusOK {
		t.Fatalf("accept inbound: %d: %s", rec.Code, rec.Body.String())
	}
	if rec := postJSON(t, srv, "/api/v1/inbound/"+in.ID+"/putaway", ""); rec.Code != http.StatusOK {
		t.Fatalf("putaway inbound: %d: %s", rec.Code, rec.Body.String())
	}
}

func countRows(t *testing.T, rec *httptest.ResponseRecorder) int {
	t.Helper()
	var rows []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &rows); err != nil {
		t.Fatalf("unmarshal: %v; body=%s", err, rec.Body.String())
	}
	return len(rows)
}
