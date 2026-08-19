# Bug 复现说明

## Bug 是什么

重试与后台效期任务忽略 context 取消传播：取消后仍继续重试、效期扫描停不下来。

## 如何触发

```bash
go test ./internal/backoff -run '^TestCtxDoAborts01$' -count=1
go test ./internal/scheduler -run '^TestCtxWorkerStops03$' -count=1
go test ./internal/expiry -run '^TestCtxLockAborts04$' -count=1
```

## 错误信息（修复前）

取消 context 后 `backoff.Do` 与效期 worker 仍继续运行，测试等待超时。
