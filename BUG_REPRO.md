# BUG_REPRO

## Bug 是什么
质检复核新增 review 中间态后，状态集合、转换表与放行/召回/处方销售策略全部漏同步：review 状态不合法、review 无法流转到 pass/fail、复核中批次仍可放行和召回。

## 如何触发
- policy.ValidReviewState("review") 返回 false；
- policy.CanReviewTransition("review","pass") 返回错误；
- policy.BatchReleaseRule(10,false,true) / RecallLevelRule(90,true) / RxSaleRule("rx",true,true,true) 均不拦截复核中批次。

## 错误信息
```
BatchReleaseRule(10, false, inReview=true) -> nil（应报 policy: batch under review cannot be released）
```
