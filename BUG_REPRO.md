# Bug 复现说明

## Bug 是什么

追溯查询接口对不存在的追溯码或批次号返回 500 服务器错误，而不是 404 Not Found。

根因是错误传播链断裂：`internal/lineage/service.go` 在二次包装底层 not-found 错误时用了 `%v` 而不是 `%w`，丢掉了 `platform.ErrNotFound` 哨兵；HTTP 层的错误状态映射又只识别校验类错误，导致 not-found 一律回退到 500。

## 如何触发

```bash
go run ./cmd/server
curl -i http://127.0.0.1:18090/api/v1/trace/code/0000000000000000000
curl -i http://127.0.0.1:18090/api/v1/trace/batch/missing-batch
```

## 错误信息（修复前）

```
HTTP/1.1 500 Internal Server Error
{"error":"trace: resolve serial: serial 0000000000000000000: not found"}
```

修复后应返回：

```
HTTP/1.1 404 Not Found
{"error":"serialization: serial 0000000000000000000 missing"}
```
