# BUG_REPRO

## Bug 是什么
批量停用客户时，循环中途遇到不存在的 id 不回滚已停用客户，defer 覆盖命名返回值把 not-found 错误吞成成功，接口对部分失败回 200。

## 如何触发
- 调用 customer.Service.SuspendMany 传入 [存在id, 不存在id]，返回 nil 错误且第一个客户仍被停用；
- HTTP: POST /api/v1/customers/bulk-suspend 夹带不存在的 id，接口回 200（应 404）且已停用客户未回滚。

## 错误信息
```
SuspendMany([good, missing]) -> (n=1, err=nil)（应 err=not found 且 good 回滚为 active）
```
