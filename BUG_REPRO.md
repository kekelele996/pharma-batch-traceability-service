# Bug 复现说明

## Bug 是什么

切片工具 `util.Filter`/`util.Dedupe` 原地复用输入底层数组，导致报表统计、出货分配的切片互相污染，并产生空的分配行。

## 如何触发

```bash
go test ./internal/util -run '^TestUtilFilterP31$' -count=1
go test ./internal/util -run '^TestUtilDedupeP32$' -count=1
go test ./internal/summary -run '^TestReportRejectedP33$' -count=1
go test ./internal/outbound -run '^TestPlanLinesP34$' -count=1
```

## 错误信息（修复前）

```
--- FAIL: TestUtilFilterP31
    slice_test.go: Filter mutated input
```
