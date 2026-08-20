package outcheck

import (
	"encoding/json"
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

// TestCreateOutboundHTTPR014D 通过 HTTP 创建出库单后，条目顺序必须保持请求原样。
func TestCreateOutboundHTTPR014D(t *testing.T) {
	srv := newServer()
	body := `{"customer_id":"c1","warehouse_id":"wh-1","items":[{"drug_id":"drug-b","qty":2},{"drug_id":"drug-a","qty":3}]}`
	rec := postJSON(t, srv, "/api/v1/outbound", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var created outbound.Outbound
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/outbound", nil)
	rec2 := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec2, req)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec2.Code)
	}
	var list []outbound.Outbound
	if err := json.Unmarshal(rec2.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	for _, o := range list {
		if o.ID != created.ID {
			continue
		}
		if len(o.Items) != 2 || o.Items[0].DrugID != "drug-b" {
			t.Fatalf("item order changed by create: %+v", o.Items)
		}
		return
	}
	t.Fatal("created outbound not found in list")
}
