# BUG_REPRO

## Bug 是什么
GS1 编码与校验在坏输入下返回零值且不带错误，nil 当成功沿调用链传播：上游序列化拿空 GTIN 继续生成追溯码，序列号接口对坏输入回 201、对不存在的码回 200 空对象。

## 如何触发
- GTIN14("") / SSCC("x","123") / NormalizeBatchNo("  ") / ComputeCheckDigit("abc") 均返回零值而非错误；
- HTTP: POST /api/v1/serials/generate 传坏 item_ref 回 201（应 400）；
- GET /api/v1/serials/不存在码 回 200 空对象（应 404）。

## 错误信息
```
POST /api/v1/serials/generate {"item_ref_13":"123"} -> 201 Created（应 400）
```
