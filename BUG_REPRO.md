# BUG_REPRO

## Bug 是什么
供应商/生产企业批量核验循环内忽略 ctx.Err()，调用方取消后仍把整个列表跑完；缺失 id 被静默跳过不回错误；BULK_VERIFY_TIMEOUT 环境变量未读取，超时配置恒为零。

## 如何触发
- 用已取消的 context 调用 supplier.Service.VerifyAll / manufacturer.Service.VerifyAll，核验不中止；
- 核验列表含不存在的 id，返回成功而非错误；
- 设置 BULK_VERIFY_TIMEOUT 后 config.Load() 的 BulkTimeout 仍为 0。

## 错误信息
```
VerifyAll(cancelled ctx) 返回 verified=3 而非 0（应返回 ctx.Err()）
```
