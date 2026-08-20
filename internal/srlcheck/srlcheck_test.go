package srlcheck

import (
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
	"pharma-batch-traceability-service/internal/httpapi"
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

func newServer() *httpapi.Server {
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
	return httpapi.NewServer(drugs, manufacturers, warehouses, suppliers, customers,
		batches, serials, stocks, shipments, tracer, inbounds, outbounds,
		dispatches, quarantines, recalls, expiries, audits, metrics,
		qualitySvc, coldchainSvc, notificationSvc, labelSvc, reportSvc)
}

func postJSON(t *testing.T, srv *httpapi.Server, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)
	return rec
}

// TestGenerateSerialsHTTPR018D 坏 item_ref 生成追溯码必须回 400。
func TestGenerateSerialsHTTPR018D(t *testing.T) {
	srv := newServer()
	rec := postJSON(t, srv, "/api/v1/serials/generate", `{"item_ref_13":"123","batch_id":"b1","count":2}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad item ref, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestGetSerialHTTP404R018G 查询不存在的追溯码必须回 404。
func TestGetSerialHTTP404R018G(t *testing.T) {
	srv := newServer()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/serials/0000000000000000000", nil)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}
