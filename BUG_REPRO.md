# Bug 复现说明

## Bug 是什么

生产批次的隔离/放行/拒收状态机失同步：隔离的批次无法放行、拒收的批次状态不对、召回只冻结部分仓库。

根因是 `internal/production/service.go` 的状态转换表缺失 `quarantined -> released` 边且 reject 分支写回旧状态，`internal/isolation/service.go` 的拒收写错状态，`internal/recall/service.go` 的冻结循环提前 break。

## 如何触发

```bash
go test ./internal/production -run '^TestBatchHoldB21$' -count=1
go test ./internal/production -run '^TestBatchUnblockB22$' -count=1
go test ./internal/production -run '^TestBatchDenyB23$' -count=1
go test ./internal/isolation -run '^TestIsolationDenyB24$' -count=1
go test ./internal/recall -run '^TestRecallStopB25$' -count=1
```

## 错误信息（修复前）

```
--- FAIL: TestBatchUnblockB22
    batch_test.go: expected released, got quarantined
```
