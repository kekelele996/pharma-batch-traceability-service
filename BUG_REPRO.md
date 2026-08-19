# Bug 复现说明

## Bug 是什么

标签批量生成预分配长度而非容量并重复追加、列表未截断产生空行、克隆/过滤原地复用底层数组造成共享污染。

## 如何触发

```bash
go test ./internal/label -run '^TestLabelGenerateP90$' -count=1
go test ./internal/label -run '^TestLabelListP91$' -count=1
go test ./internal/label -run '^TestLabelCloneP92$' -count=1
go test ./internal/label -run '^TestLabelFilterP93$' -count=1
```

## 错误信息（修复前）

标签数量与请求不符、列表带空行、克隆/过滤修改原切片。
