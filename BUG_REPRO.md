# Bug 复现说明

## Bug 是什么

采购订单状态机缺失合法状态与转换守卫：草稿单可直接下单、收货写错状态、已收货单可被取消。

## 如何触发

```bash
go test ./internal/po -run '^TestPoTransP71$' -count=1
go test ./internal/po -run '^TestPoOrderP72$' -count=1
go test ./internal/po -run '^TestPoReceiP73$' -count=1
go test ./internal/po -run '^TestPoCanceP74$' -count=1
```

## 错误信息（修复前）

```
--- FAIL: TestPoOrderP72
    po_test.go: expected ordering an unapproved po to fail
```
