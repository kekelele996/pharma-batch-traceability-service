package httpapi

import (
	"net/http"

	"pharma-batch-traceability-service/internal/audit"
	"pharma-batch-traceability-service/internal/production"
	"pharma-batch-traceability-service/internal/coldchain"
	"pharma-batch-traceability-service/internal/customer"
	"pharma-batch-traceability-service/internal/relocation"
	"pharma-batch-traceability-service/internal/drug"
	"pharma-batch-traceability-service/internal/expiry"
	"pharma-batch-traceability-service/internal/inbound"
	"pharma-batch-traceability-service/internal/label"
	"pharma-batch-traceability-service/internal/manufacturer"
	"pharma-batch-traceability-service/internal/metric"
	"pharma-batch-traceability-service/internal/notification"
	"pharma-batch-traceability-service/internal/outbound"
	"pharma-batch-traceability-service/internal/platform"
	"pharma-batch-traceability-service/internal/quality"
	"pharma-batch-traceability-service/internal/isolation"
	"pharma-batch-traceability-service/internal/recall"
	"pharma-batch-traceability-service/internal/report"
	"pharma-batch-traceability-service/internal/serialization"
	"pharma-batch-traceability-service/internal/shipment"
	"pharma-batch-traceability-service/internal/stock"
	"pharma-batch-traceability-service/internal/supplier"
	"pharma-batch-traceability-service/internal/lineage"
	"pharma-batch-traceability-service/internal/warehouse"
)

type Server struct {
	drugs         *drug.Service
	manufacturers *manufacturer.Service
	warehouses    *warehouse.Service
	suppliers     *supplier.Service
	customers     *customer.Service
	batches       *production.Service
	serials       *serialization.Service
	stock         *stock.Service
	shipments     *shipment.Service
	trace         *lineage.Service
	inbound       *inbound.Service
	outbound      *outbound.Service
	dispatch      *relocation.Service
	quarantine    *isolation.Service
	recall        *recall.Service
	expiry        *expiry.Service
	audit         *audit.Service
	metric        *metric.Service
	quality       *quality.Service
	coldchain     *coldchain.Service
	notification  *notification.Service
	label         *label.Service
	report        *report.Service
}

func NewServer(d *drug.Service, m *manufacturer.Service, w *warehouse.Service, sup *supplier.Service,
	c *customer.Service, b *production.Service, srl *serialization.Service, st *stock.Service,
	sh *shipment.Service, tr *lineage.Service, in *inbound.Service, out *outbound.Service,
	dsp *relocation.Service, qa *isolation.Service, rc *recall.Service, ex *expiry.Service,
	au *audit.Service, me *metric.Service, qual *quality.Service, cc *coldchain.Service,
	ntf *notification.Service, lbl *label.Service, rpt *report.Service) *Server {
	return &Server{
		drugs: d, manufacturers: m, warehouses: w, suppliers: sup, customers: c,
		batches: b, serials: srl, stock: st, shipments: sh, trace: tr,
		inbound: in, outbound: out, dispatch: dsp, quarantine: qa, recall: rc,
		expiry: ex, audit: au, metric: me, quality: qual, coldchain: cc,
		notification: ntf, label: lbl, report: rpt,
	}
}

func (s *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /api/v1/metrics", s.metrics)
	mux.HandleFunc("GET /api/v1/report/summary", s.reportSummary)

	mux.HandleFunc("POST /api/v1/drugs", s.createDrug)
	mux.HandleFunc("GET /api/v1/drugs", s.listDrugs)
	mux.HandleFunc("GET /api/v1/drugs/{id}", s.getDrug)
	mux.HandleFunc("PUT /api/v1/drugs/{id}", s.updateDrug)

	mux.HandleFunc("POST /api/v1/manufacturers", s.createManufacturer)
	mux.HandleFunc("GET /api/v1/manufacturers", s.listManufacturers)
	mux.HandleFunc("POST /api/v1/manufacturers/{id}/status", s.setManufacturerStatus)

	mux.HandleFunc("POST /api/v1/warehouses", s.createWarehouse)
	mux.HandleFunc("GET /api/v1/warehouses", s.listWarehouses)
	mux.HandleFunc("POST /api/v1/warehouses/{id}/status", s.setWarehouseStatus)

	mux.HandleFunc("POST /api/v1/suppliers", s.createSupplier)
	mux.HandleFunc("GET /api/v1/suppliers", s.listSuppliers)

	mux.HandleFunc("POST /api/v1/customers", s.createCustomer)
	mux.HandleFunc("GET /api/v1/customers", s.listCustomers)
	mux.HandleFunc("POST /api/v1/customers/{id}/rx-permit", s.setCustomerRxPermit)

	mux.HandleFunc("POST /api/v1/batches", s.createBatch)
	mux.HandleFunc("GET /api/v1/batches", s.listBatches)
	mux.HandleFunc("GET /api/v1/batches/{id}", s.getBatch)
	mux.HandleFunc("POST /api/v1/batches/{id}/release", s.releaseBatch)
	mux.HandleFunc("POST /api/v1/batches/{id}/reject", s.rejectBatch)
	mux.HandleFunc("POST /api/v1/batches/{id}/quarantine", s.quarantineBatch)

	mux.HandleFunc("POST /api/v1/serials/generate", s.generateSerials)
	mux.HandleFunc("GET /api/v1/serials/{code}", s.getSerial)
	mux.HandleFunc("POST /api/v1/serials/{code}/activate", s.activateSerial)
	mux.HandleFunc("POST /api/v1/serials/{code}/sell", s.sellSerial)
	mux.HandleFunc("POST /api/v1/serials/{code}/void", s.voidSerial)
	mux.HandleFunc("GET /api/v1/serials", s.listSerials)

	mux.HandleFunc("GET /api/v1/stock", s.listStock)

	mux.HandleFunc("POST /api/v1/inbound", s.createInbound)
	mux.HandleFunc("GET /api/v1/inbound", s.listInbound)
	mux.HandleFunc("POST /api/v1/inbound/{id}/accept", s.acceptInbound)
	mux.HandleFunc("POST /api/v1/inbound/{id}/putaway", s.putawayInbound)

	mux.HandleFunc("POST /api/v1/outbound", s.createOutbound)
	mux.HandleFunc("GET /api/v1/outbound", s.listOutbound)
	mux.HandleFunc("POST /api/v1/outbound/{id}/allocate", s.allocateOutbound)
	mux.HandleFunc("POST /api/v1/outbound/{id}/ship", s.shipOutbound)
	mux.HandleFunc("POST /api/v1/outbound/{id}/cancel", s.cancelOutbound)

	mux.HandleFunc("POST /api/v1/dispatch", s.createDispatch)
	mux.HandleFunc("GET /api/v1/dispatch", s.listDispatch)
	mux.HandleFunc("POST /api/v1/dispatch/{id}/start", s.startDispatch)
	mux.HandleFunc("POST /api/v1/dispatch/{id}/complete", s.completeDispatch)
	mux.HandleFunc("POST /api/v1/dispatch/{id}/cancel", s.cancelDispatch)

	mux.HandleFunc("POST /api/v1/quarantine", s.placeQuarantine)
	mux.HandleFunc("GET /api/v1/quarantine", s.listQuarantine)
	mux.HandleFunc("POST /api/v1/quarantine/{id}/release", s.releaseQuarantine)
	mux.HandleFunc("POST /api/v1/quarantine/{id}/reject", s.rejectQuarantine)

	mux.HandleFunc("POST /api/v1/recall", s.issueRecall)
	mux.HandleFunc("GET /api/v1/recall", s.listRecall)
	mux.HandleFunc("GET /api/v1/recall/{id}/list", s.recallList)
	mux.HandleFunc("POST /api/v1/recall/{id}/execute", s.executeRecall)
	mux.HandleFunc("POST /api/v1/recall/{id}/complete", s.completeRecall)

	mux.HandleFunc("GET /api/v1/expiry/scan", s.scanExpiry)
	mux.HandleFunc("POST /api/v1/expiry/lock", s.lockExpired)

	mux.HandleFunc("GET /api/v1/trace/code/{code}", s.traceByCode)
	mux.HandleFunc("GET /api/v1/trace/batch/{batchID}", s.traceByBatch)

	mux.HandleFunc("POST /api/v1/quality", s.createQuality)
	mux.HandleFunc("GET /api/v1/quality", s.listQuality)
	mux.HandleFunc("POST /api/v1/quality/{id}/decide", s.decideQuality)

	mux.HandleFunc("POST /api/v1/coldchain/logs", s.recordColdchain)
	mux.HandleFunc("GET /api/v1/coldchain/logs", s.listColdchain)

	mux.HandleFunc("POST /api/v1/notifications", s.enqueueNotification)
	mux.HandleFunc("POST /api/v1/notifications/send", s.sendNotifications)
	mux.HandleFunc("GET /api/v1/notifications", s.listNotifications)

	mux.HandleFunc("POST /api/v1/labels/generate", s.generateLabels)
	mux.HandleFunc("POST /api/v1/labels/{id}/print", s.printLabel)
	mux.HandleFunc("GET /api/v1/labels", s.listLabels)

	mux.HandleFunc("GET /api/v1/audit", s.listAudit)
	return mux
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	platform.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) metrics(w http.ResponseWriter, r *http.Request) {
	platform.WriteJSON(w, http.StatusOK, s.metric.Snapshot())
}

func writeErr(w http.ResponseWriter, err error) {
	switch {
	case platform.IsNotFound(err):
		platform.WriteError(w, http.StatusNotFound, err.Error())
	case platform.IsConflict(err):
		platform.WriteError(w, http.StatusConflict, err.Error())
	case platform.IsValidation(err):
		platform.WriteError(w, http.StatusBadRequest, err.Error())
	default:
		platform.WriteError(w, http.StatusInternalServerError, err.Error())
	}
}

func pathID(r *http.Request, name string) string { return r.PathValue(name) }
