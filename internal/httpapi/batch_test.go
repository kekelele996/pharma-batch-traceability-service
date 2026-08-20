package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"pharma-batch-traceability-service/internal/production"
)

func newBatchID(t *testing.T, srv *Server) string {
	t.Helper()
	body := `{"drug_id":"d1","batch_no":"BT-CONC-1","production_date":"2026-07-01T00:00:00Z","expiry_date":"2027-07-01T00:00:00Z","quantity":100}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/batches", strings.NewReader(body))
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create batch: expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var b production.Batch
	if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil {
		t.Fatal(err)
	}
	return b.ID
}

// TestBatchConcurrentReleaseP13 并发放行同一批次不能出现 concurrent map write。
func TestBatchConcurrentReleaseR017C(t *testing.T) {
	srv := newTestServer()
	id := newBatchID(t, srv)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			req := httptest.NewRequest(http.MethodPost, "/api/v1/batches/"+id+"/release", nil)
			rec := httptest.NewRecorder()
			srv.Routes().ServeHTTP(rec, req)
		}()
	}
	close(start)
	wg.Wait()
}
