# pharma-batch-traceability-service

药品批次追溯平台（Go + 标准库 + 前端查询页）。围绕药品从生产批次、电子追溯码、出入库、调拨、隔离放行、召回到效期管理的全链路，提供一套可运行、可追溯的 B2B 追溯服务。

## 目录结构

```
cmd/server            服务入口（HTTP 路由 + 演示数据 seed）
internal/platform     共享基础设施（错误、JSON、ID、时钟）
internal/gs1          GS1 工具（GTIN-14 校验位、SSCC、批号规范化）
internal/drug         药品档案
internal/manufacturer 生产企业
internal/warehouse    仓库与温区
internal/supplier     供应商
internal/customer     经销商 / 客户
internal/batch        生产批次（待检/隔离/放行/拒收状态机）
internal/serialization 电子追溯码（GTIN14 + 序列号）
internal/stock        批次库存（入库/出库/锁定/冻结）
internal/inbound      入库单
internal/outbound     出库单（FEFO 分配）
internal/dispatch     跨库调拨
internal/quarantine   隔离放行
internal/recall       召回
internal/expiry       效期管理
internal/trace        追溯链聚合
internal/policy       业务规则（储存/处方/冷链）
internal/audit        审计日志
internal/metric       指标统计
internal/config       配置
internal/httpapi       HTTP 处理器与路由
web                   追溯查询前端页
```

## 运行

```bash
go run ./cmd/server
# 服务默认监听 :18090，健康检查 http://127.0.0.1:18090/health
# 前端查询页 http://127.0.0.1:18090/
```

环境变量：

- `PORT`：监听端口，默认 `18090`
- `APP_ENV`：运行环境，默认 `dev`

## 测试

```bash
go test ./...
```

## 主要 API

- `POST /api/v1/batches` 创建生产批次；`POST /api/v1/batches/{id}/release` 放行
- `POST /api/v1/serials/generate` 为批次生成电子追溯码
- `POST /api/v1/inbound` / `POST /api/v1/inbound/{id}/putaway` 入库与上架
- `POST /api/v1/outbound` / `allocate` / `ship` 出库与 FEFO 分配
- `GET /api/v1/trace/code/{code}` 按追溯码查询全链路
- `POST /api/v1/recall` 发起召回；`POST /api/v1/quarantine` 隔离
- `GET /api/v1/expiry/scan` 效期扫描
