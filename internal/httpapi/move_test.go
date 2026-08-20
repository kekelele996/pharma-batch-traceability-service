package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"pharma-batch-traceability-service/internal/relocation"
)

func newDispatchID(t *testing.T, srv *Server) string {
	t.Helper()
	body := `{"from_warehouse_id":"wh-a","to_warehouse_id":"wh-b","items":[{"drug_id":"d1","batch_id":"b1","qty":2}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/dispatch", strings.NewReader(body))
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create dispatch: expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var d relocation.Dispatch
	if err := json.Unmarshal(rec.Body.Bytes(), &d); err != nil {
		t.Fatal(err)
	}
	return d.ID
}

// TestDispatchConcurrentStartP11 并发 start 同一调拨不能出现 concurrent map write。
func TestDispatchConcurrentStartR017A(t *testing.T) {
	srv := newTestServer()
	id := newDispatchID(t, srv)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			req := httptest.NewRequest(http.MethodPost, "/api/v1/dispatch/"+id+"/start", nil)
			rec := httptest.NewRecorder()
			srv.Routes().ServeHTTP(rec, req)
		}()
	}
	close(start)
	wg.Wait()
}


// TestDispatchConcurrentCompleteP12 并发 complete/cancel 已开始的同一调拨不能出现 data race。
func TestDispatchConcurrentCompleteR017B(t *testing.T) {
	srv := newTestServer()
	id := newDispatchID(t, srv)
	dispatchOps[id] = true
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			req := httptest.NewRequest(http.MethodPost, "/api/v1/dispatch/"+id+"/complete", nil)
			rec := httptest.NewRecorder()
			srv.Routes().ServeHTTP(rec, req)
		}()
	}
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			req := httptest.NewRequest(http.MethodPost, "/api/v1/dispatch/"+id+"/cancel", nil)
			rec := httptest.NewRecorder()
			srv.Routes().ServeHTTP(rec, req)
		}()
	}
	close(start)
	wg.Wait()
	delete(dispatchOps, id)
}

