# Bug 复现说明

## Bug 是什么

调拨状态机允许越级与回退：未发车可完成、已完成可取消、同仓可建调拨。

## 如何触发

```bash
go test ./internal/relocation -run '^TestRelocationTransA1$' -count=1
go test ./internal/relocation -run '^TestRelocationCompleteA2$' -count=1
go test ./internal/relocation -run '^TestRelocationCancelA3$' -count=1
go test ./internal/relocation -run '^TestRelocationSameA4$' -count=1
```

## 错误信息（修复前）

未发车调拨可直接完成、已完成可取消、同仓建单成功。
