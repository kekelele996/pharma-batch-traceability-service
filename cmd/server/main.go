package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pharma-batch-traceability-service/internal/audit"
	"pharma-batch-traceability-service/internal/production"
	"pharma-batch-traceability-service/internal/coldchain"
	"pharma-batch-traceability-service/internal/config"
	"pharma-batch-traceability-service/internal/customer"
	"pharma-batch-traceability-service/internal/relocation"
	"pharma-batch-traceability-service/internal/drug"
	"pharma-batch-traceability-service/internal/expiry"
	"pharma-batch-traceability-service/internal/httpapi"
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
	"pharma-batch-traceability-service/internal/summary"
	"pharma-batch-traceability-service/internal/serialization"
	"pharma-batch-traceability-service/internal/shipment"
	"pharma-batch-traceability-service/internal/stock"
	"pharma-batch-traceability-service/internal/supplier"
	"pharma-batch-traceability-service/internal/lineage"
	"pharma-batch-traceability-service/internal/warehouse"
	"pharma-batch-traceability-service/internal/scheduler"
)

func main() {
	cfg := config.Load()
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

	seed(clock, drugs, warehouses, batches, serials, stocks, shipments)

	server := httpapi.NewServer(drugs, manufacturers, warehouses, suppliers, customers,
		batches, serials, stocks, shipments, tracer, inbounds, outbounds,
		dispatches, quarantines, recalls, expiries, audits, metrics,
		qualitySvc, coldchainSvc, notificationSvc, labelSvc, reportSvc)

	mux := server.Routes()
	mux.Handle("/", http.FileServer(http.Dir("web")))

	expiryWorker := scheduler.NewExpiryWorker(time.Minute, expiries)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	expiryWorker.Start(ctx)

	addr := ":" + cfg.Port
	srv := &http.Server{Addr: addr, Handler: mux}
	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()
	log.Printf("pharma batch traceability service listening on %s (env=%s)", addr, cfg.Env)

	select {
	case err := <-errCh:
		log.Fatalf("serve: %v", err)
	case <-ctx.Done():
		log.Println("shutdown signal received, stopping")
		_ = srv.Close()
		expiryWorker.Wait()
	}
}

// seed registers a small demo dataset so the trace page and metrics have
// something to show on first boot.
func seed(clock platform.Clock, drugs *drug.Service, whs *warehouse.Service, batches *production.Service,
	serials *serialization.Service, stocks *stock.Service, shipments *shipment.Service) {
	wh, err := whs.Create(warehouse.Warehouse{
		Code: "WH-DEMO", Name: "演示仓", TempZone: "cold", Province: "上海", City: "上海", Status: "active",
	})
	if err != nil {
		log.Printf("seed warehouse: %v", err)
		return
	}
	d, err := drugs.Create(drug.Drug{
		Code: "H20240001", GenericName: "注射用头孢曲松钠", TradeName: "罗氏芬",
		DosageForm: "injection", Strength: "1.0g", RxCategory: "rx", Storage: "cold", ManufacturerID: "MFR-DEMO",
	})
	if err != nil {
		log.Printf("seed drug: %v", err)
		return
	}
	b, err := batches.Create(production.Batch{
		DrugID: d.ID, BatchNo: "BT202408001",
		ProductionDate: time.Now().UTC().AddDate(0, -1, 0),
		ExpiryDate:     time.Now().UTC().AddDate(0, 11, 0),
		Quantity:       1000,
	})
	if err != nil {
		log.Printf("seed batch: %v", err)
		return
	}
	if _, err := batches.Release(b.ID); err != nil {
		log.Printf("seed release: %v", err)
		return
	}
	if _, err := stocks.Receive(b.ID, wh.ID, 1000); err != nil {
		log.Printf("seed stock: %v", err)
		return
	}
	if _, err := shipments.Append(shipment.Shipment{
		BatchID: b.ID, FromID: "MFR-DEMO", ToID: wh.ID, NodeType: shipment.NodeProduction,
		OccurredAt: b.ProductionDate, TempC: 4,
	}); err != nil {
		log.Printf("seed shipment: %v", err)
	}
	if items, err := serials.Generate("6901234567890", b.ID, 5); err != nil {
		log.Printf("seed serials: %v", err)
	} else {
		log.Printf("seeded %d serials for demo batch %s", len(items), b.ID)
	}
	log.Printf("seed complete: drug=%s batch=%s warehouse=%s", d.ID, b.ID, wh.ID)
}
