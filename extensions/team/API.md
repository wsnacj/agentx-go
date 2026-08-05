# Team Topology API（Experimental）

`extensions/team`定义可移植Team拓扑和确定性协作阶段，不执行成员，也不拥有scheduler。

- `Spec`包含Team identity、coordinator和引用Expert ID的Members；
- `Member.DependsOn`形成DAG，`Normalize`验证引用、重复、self-dependency和cycle；
- `BuildPlan`执行稳定topological planning，同一Stage成员彼此无依赖并按ID排序；
- `Plan`只含Team、Coordinator、Stage和Member，不含worker、queue、model、budget、retry或状态；
- `Project`只生成display-safe `catalog.KindTeam`资产，不暴露拓扑和责任正文；
- `Parse`拒绝scheduler、concurrency、model、credential等Host-owned字段。

同一Stage是否并行、如何实例化Session/Subagent、预算/审批/取消/失败恢复和成员输出验证，都由
Platform Target复用现有Runtime owner决定。本包不是第二套Agent Loop或多Agent scheduler。

当前为Experimental，不构成Public/Beta/Stable或生产编排承诺。
