---
description: "
源码分析 OpenKruise，并从中来看 Controller 的设计，并提出自己的思考
"
---

面向状态的编程的难点在于处理函数要幂等地处理对象的所有状态。
处理状态时候对先后顺序要特别讲究？

核心设计
1、watch 机制
2、Expectation 期望和观察确认机制
3、延迟入队列机制
4、幂等设计

：
1、需要 watch 哪些东西？
2、controller-runtimne 使用的介绍
3、如何 CRD 处理新建、更新、删除的场景的？

Advancedcronjob

reconcileJob
    1. 根据 NamespaceName 获取 Job
    2. 根据 Jobs 的统计更新 ACJ 状态
    3. 新建场景，更新失败，报错 notfound

broadcastjob
1. watch BCJ 的更新、POD 的更新、Node 的更新
2. 是否匹配 satisfiedExpectation 并配置 requeue 的时间
3. 判断 BCJ 是否结束，并删除 BCJ
4. 通过 StartTime 判断 BCJ 是否新建，这个时候配置当前时间为 startTime，然后 now+TTL 为 requeue 的时间，配置 Status.Phase 为 running
遍历所有所属 BCJ 的 Pod，然后计算 Pod 状态、还需要在多少个 Node 创建 POD、重置 Pod、删除 Pod
   
针对 BCJ 的 failed、paused 更新 BCJ Status
根据 policy 更新 BCJ status
如果 Failed 失败，删除正在运行中的 POD，否则处理多余的 POD，创建期望的 POD
更新 BCJ 的 Status