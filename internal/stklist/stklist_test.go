package stklist

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

func post(t *testing.T, srv *httpapi.Server, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)
	return rec
}

func get(t *testing.T, srv *httpapi.Server, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)
	return rec
}

func seedTwoStockRows(t *testing.T, srv *httpapi.Server) {
	t.Helper()
	ids := []string{}
	for _, no := range []string{"BTA", "BTB"} {
		body := fmt.Sprintf(`{"drug_id":"d1","batch_no":"%s","production_date":"2026-07-01T00:00:00Z","expiry_date":"2027-07-01T00:00:00Z","quantity":100}`, no)
		rec := post(t, srv, "/api/v1/batches", body)
		if rec.Code != http.StatusCreated {
			t.Fatalf("create batch: %d: %s", rec.Code, rec.Body.String())
		}
		var b production.Batch
		if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, b.ID)
		if rec := post(t, srv, "/api/v1/batches/"+b.ID+"/release", ""); rec.Code != http.StatusOK {
			t.Fatalf("release: %d: %s", rec.Code, rec.Body.String())
		}
	}
	inBody := fmt.Sprintf(`{"supplier_id":"s1","warehouse_id":"wh-1","items":[{"drug_id":"d1","batch_id":"%s","qty":10},{"drug_id":"d1","batch_id":"%s","qty":20}]}`, ids[0], ids[1])
	rec := post(t, srv, "/api/v1/inbound", inBody)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create inbound: %d: %s", rec.Code, rec.Body.String())
	}
	var in inbound.Inbound
	if err := json.Unmarshal(rec.Body.Bytes(), &in); err != nil {
		t.Fatal(err)
	}
	if rec := post(t, srv, "/api/v1/inbound/"+in.ID+"/accept?qc_result=pass", ""); rec.Code != http.StatusOK {
		t.Fatalf("accept: %d: %s", rec.Code, rec.Body.String())
	}
	if rec := post(t, srv, "/api/v1/inbound/"+in.ID+"/putaway", ""); rec.Code != http.StatusOK {
		t.Fatalf("putaway: %d: %s", rec.Code, rec.Body.String())
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

// TestListStockNoFilterR019A 带过滤查询之后，不带过滤的库存列表必须返回全部行。
func TestListStockNoFilterR019A(t *testing.T) {
	srv := newServer()
	seedTwoStockRows(t, srv)
	if rec := get(t, srv, "/api/v1/stock?batch_id=no-such-batch"); rec.Code != http.StatusOK {
		t.Fatalf("filtered stock list: %d", rec.Code)
	}
	recAll := get(t, srv, "/api/v1/stock")
	if n := countRows(t, recAll); n != 2 {
		t.Fatalf("expected 2 stock rows without filter, got %d: %s", n, recAll.Body.String())
	}
}

// TestListInboundNoFilterR019B 带状态过滤之后，不带过滤的入库列表必须返回全部。
func TestListInboundNoFilterR019B(t *testing.T) {
	srv := newServer()
	post(t, srv, "/api/v1/inbound", `{"supplier_id":"s1","warehouse_id":"wh-1","items":[{"drug_id":"d1","batch_id":"bA","qty":1}]}`)
	recB := post(t, srv, "/api/v1/inbound", `{"supplier_id":"s1","warehouse_id":"wh-1","items":[{"drug_id":"d1","batch_id":"bB","qty":1}]}`)
	var b inbound.Inbound
	if err := json.Unmarshal(recB.Body.Bytes(), &b); err != nil {
		t.Fatal(err)
	}
	if rec := post(t, srv, "/api/v1/inbound/"+b.ID+"/accept?qc_result=pass", ""); rec.Code != http.StatusOK {
		t.Fatalf("accept B: %d: %s", rec.Code, rec.Body.String())
	}
	if rec := get(t, srv, "/api/v1/inbound?status=draft"); rec.Code != http.StatusOK {
		t.Fatalf("filtered inbound list: %d", rec.Code)
	}
	recAll := get(t, srv, "/api/v1/inbound")
	if n := countRows(t, recAll); n != 2 {
		t.Fatalf("expected 2 inbound rows without filter, got %d: %s", n, recAll.Body.String())
	}
}

// TestListStockFilterIsolatedR019C 带过滤查询的结果不能混入上一次查询的内容。
func TestListStockFilterIsolatedR019C(t *testing.T) {
	srv := newServer()
	seedTwoStockRows(t, srv)
	// find an existing batch id from the unfiltered list
	all := get(t, srv, "/api/v1/stock")
	var rows []map[string]any
	if err := json.Unmarshal(all.Body.Bytes(), &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) == 0 {
		t.Fatal("expected seeded stock rows")
	}
	batchID := rows[0]["batch_id"].(string)
	if rec := get(t, srv, "/api/v1/stock?batch_id="+batchID); rec.Code != http.StatusOK {
		t.Fatalf("batch filtered stock list: %d", rec.Code)
	}
	rec := get(t, srv, "/api/v1/stock?warehouse_id=wh-2")
	if n := countRows(t, rec); n != 0 {
		t.Fatalf("expected 0 rows for wh-2, got %d: %s", n, rec.Body.String())
	}
}

// TestListInboundFilterIsolatedR019D 状态过滤的结果不能混入上一次查询的内容。
func TestListInboundFilterIsolatedR019D(t *testing.T) {
	srv := newServer()
	post(t, srv, "/api/v1/inbound", `{"supplier_id":"s1","warehouse_id":"wh-1","items":[{"drug_id":"d1","batch_id":"bA","qty":1}]}`)
	recB := post(t, srv, "/api/v1/inbound", `{"supplier_id":"s1","warehouse_id":"wh-1","items":[{"drug_id":"d1","batch_id":"bB","qty":1}]}`)
	var b inbound.Inbound
	if err := json.Unmarshal(recB.Body.Bytes(), &b); err != nil {
		t.Fatal(err)
	}
	if rec := post(t, srv, "/api/v1/inbound/"+b.ID+"/accept?qc_result=pass", ""); rec.Code != http.StatusOK {
		t.Fatalf("accept B: %d: %s", rec.Code, rec.Body.String())
	}
	if rec := get(t, srv, "/api/v1/inbound?status=draft"); rec.Code != http.StatusOK {
		t.Fatalf("draft filter: %d", rec.Code)
	}
	rec := get(t, srv, "/api/v1/inbound?status=accepted")
	if n := countRows(t, rec); n != 1 {
		t.Fatalf("expected 1 accepted inbound, got %d: %s", n, rec.Body.String())
	}
}
