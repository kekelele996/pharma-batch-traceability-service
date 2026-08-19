# Bug 复现说明

## Bug 是什么

指标维度计数与通知渠道索引的写路径未初始化内层 map，首次写入触发 nil map panic；Dims/Snapshot 返回内部 map 造成引用逃逸。

## 如何触发

```bash
go test ./internal/metric -run '^TestPbtcIncDimNoPanic$' -count=1
go test ./internal/notification -run '^TestEnqueueNoPanic04$' -count=1
```

## 错误信息（修复前）

```
panic: assignment to entry in nil map [recovered, repanicked]

goroutine 21 [running]:
  internal/metric.(*Service).IncDim(...)
      internal/metric/service.go:41 +0x90
```
