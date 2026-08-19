# Bug 复现说明

## Bug 是什么

入库上架循环吞错断链：批次缺失/未放行错误被吞导致部分上架且无回滚；创建单忽略校验错误。

## 如何触发

```bash
go test ./internal/inbound -run '^TestInboundRollbackP81$' -count=1
go test ./internal/inbound -run '^TestInboundMissingP82$' -count=1
go test ./internal/inbound -run '^TestInboundCreateValidateP83$' -count=1
```

## 错误信息（修复前）

部分上架后未回滚，缺失批次/未放行批次未返回错误。
