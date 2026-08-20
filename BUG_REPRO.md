# BUG_REPRO

## Bug 是什么
温度记录与设备校准的读路径不加读锁，审计列表返回内部切片引用。并发读写温度/设备数据时触发 data race，读取者拿到的列表快照会被后续写入污染。

## 如何触发
- 并发执行温度记录（coldchain.Record）与读取最值/列表（Min / ListByShipment / List），`go test -race` 报 data race；
- 并发执行设备校准（device.Calibrate）与列表读取（List / CalibrationDue），`go test -race` 报 data race；
- 审计列表 List 返回内部切片，调用方修改返回值会污染内部记录。

## 错误信息
```
WARNING: DATA RACE
Read at 0x... by goroutine N:
  pharma-batch-traceability-service/internal/coldchain.(*Service).ListByShipment()
Write at 0x... by goroutine M:
  pharma-batch-traceability-service/internal/coldchain.(*Service).Record()
```
