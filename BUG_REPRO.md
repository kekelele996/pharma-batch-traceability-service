# BUG_REPRO

## Bug 是什么
仓库状态更新时向未初始化的 statusIndex nil map 写入，触发 assignment to entry in nil map panic；按状态查询恒为空。另有缺失仓库/药品查询返回零值当成功、空温区查询返回 ok=true 的 typed-nil、非法药品创建静默成功等零值路径问题。

## 如何触发
- 首次调用 warehouse.SetStatus 更新仓库状态（HTTP: POST /api/v1/warehouses/{id}/status），进程 panic；
- 查询不存在的仓库/药品存储条件返回成功；
- 空温区查询 LatestByZone 返回 ok=true。

## 错误信息
```
panic: assignment to entry in nil map [recovered, repanicked]
goroutine 7 [running]:
  pharma-batch-traceability-service/internal/warehouse.(*Service).SetStatus()
      internal/warehouse/service.go:45 +0x534
```
