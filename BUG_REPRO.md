# BUG_REPRO

## Bug 是什么
出库分配复用 service 级 scratch 切片共享底层数组，后一张出库单覆盖前一张的分配明细；Create 不拷贝 Items，Validate 与 HTTP handler 原地改写调用方切片，导致单据数据串场、调用方输入被串改。

## 如何触发
- 连续对两张出库单执行 Allocate，第一张的 Allocations 被第二张覆盖；
- 创建单据后修改调用方持有的 Items 切片，单据内容被串改；
- Validate / createOutbound 会原地排序改写传入切片。

## 错误信息
```
order A allocations corrupted after allocating order B: [{B1 20}]（应保持 [{A1 10}]）
```
