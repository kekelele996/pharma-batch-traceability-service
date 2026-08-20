# BUG_REPRO

## Bug 是什么
httpapi 调拨/放行 handler 的去重 map 无锁读写，并发 start/complete/cancel 同一调拨或并发放行批次时触发 concurrent map write / data race。

## 如何触发
- 两个请求同时 POST /api/v1/dispatch/{id}/start（或 complete/cancel 同一调拨），服务 panic；
- 两个请求同时 POST /api/v1/batches/{id}/release，`go test -race` 报 data race。

## 错误信息
```
WARNING: DATA RACE
Read at 0x... by goroutine 10:
  pharma-batch-traceability-service/internal/httpapi.(*Server).startDispatch()
      internal/httpapi/move.go:33 +0x140
Write at 0x... by goroutine 11:
  pharma-batch-traceability-service/internal/httpapi.(*Server).startDispatch()
      internal/httpapi/move.go:37 +0x17c
```
