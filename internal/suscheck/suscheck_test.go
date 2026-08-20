package suscheck

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

// TestBulkSuspendCustomersHTTPR016D 批量停用接口遇到不存在的客户必须回 4xx 且已停用客户回滚。
func TestBulkSuspendCustomersHTTPR016D(t *testing.T) {
	srv := newServer()
	create := postJSON(t, srv, "/api/v1/customers", `{"name":"cust-x","credit_code":"913100001234567890","license_no":"L","status":"active"}`)
	if create.Code != http.StatusCreated {
		t.Fatalf("create customer: expected 201, got %d", create.Code)
	}
	var created customer.Customer
	if err := json.Unmarshal(create.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	ids, _ := json.Marshal([]string{created.ID, "missing-cust"})
	rec := postJSON(t, srv, "/api/v1/customers/bulk-suspend", `{"ids":`+string(ids)+`}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing customer, got %d: %s", rec.Code, rec.Body.String())
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/customers", nil)
	rec2 := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec2, req)
	var list []customer.Customer
	if err := json.Unmarshal(rec2.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	for _, c := range list {
		if c.ID == created.ID {
			if c.Status != "active" {
				t.Fatalf("customer should remain active after failed bulk suspend, got %q", c.Status)
			}
			return
		}
	}
	t.Fatal("created customer not found in list")
}
