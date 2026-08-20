# BUG_REPRO

## Bug 是什么
质量/检查决策接口在记录不存在或重复时用 %v 包装错误，errors.Is(NotFound/Conflict) 失效；httpapi 决策接口把 not-found 与非法结果全部映射成 500。

## 如何触发
- 对不存在的质量记录/检查记录执行决策（Decide），接口回 500（应 404）；
- 提交非法结果值时接口回 500（应 400）；
- 重复 id 创建检查记录时静默覆盖而非报冲突。

## 错误信息
```
POST /api/v1/quality/{id}/decide -> 500 Internal Server Error
（基线断言 errors.Is(err, ErrNotFound) 为 false）
```
