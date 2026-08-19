# Bug 复现说明

## Bug 是什么

库存扣减/锁定/解锁读路径不加锁造成 check-then-act 竞态（data race）与超卖；shipment 追加不加锁且吞校验、重复入列。

## 如何触发

```bash
go test -race ./internal/stock -run '^TestConcDeductRace01$' -count=1
go test -race ./internal/stock -run '^TestConcLockRace03$' -count=1
go test -race ./internal/stock -run '^TestConcUnlockRace04$' -count=1
go test -race ./internal/shipment -run '^TestConcShipmentAppendRace$' -count=1
```

## 错误信息（修复前）

```
WARNING: DATA RACE
Write at 0x00c00007d3b0 by goroutine 15:
  pharma-batch-traceability-service/internal/stock.(*Service).Deduct()
      internal/stock/service.go:63 +0x3d4
Previous read at 0x00c00007d3b0 by goroutine 11:
  pharma-batch-traceability-service/internal/stock.(*Service).get()
      internal/stock/service.go:20 +0x1bc
```
