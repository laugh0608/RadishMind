# Workflow Draft Designer Surface 专题

更新时间：2026-06-16

## 专题定位

`Workflow Draft Designer Surface` 是 `apps/radishmind-web/` 中承载 workflow 草案查看、模板切换、节点 / 边审查、风险摘要和 blocked capability 展示的页面专题。

它不是 workflow builder 完整实现，也不是 workflow executor。当前页面的核心职责是把草案设计状态组织成可审查信息，并通过 `Workflow Draft Editing Entry v1`、`Workflow Draft Designer Editing Model v2` 和 `Workflow Draft Node Attribute Editing Model v1` 提供受控本地编辑入口、本地图结构编辑能力和节点属性编辑能力。

## 当前状态

- 当前 surface 以离线 view model 和 workflow context 作为初始数据来源。
- 已展示 draft templates、本地 template switch、draft nodes、edges、readiness、risk summary、route / request / audit metadata 和 blocked capability preview。
- 当前已可在显式 dev-only consumer 配置下保存、读取和校验 saved draft，并在页面上区分 sample / unsaved / saved / failed 状态。
- 当前已提供受控本地编辑入口，可编辑草案名称、说明、节点名称和边条件摘要；validate / save / read 使用当前本地草案。
- 当前已接入 `User Workspace Draft Creation v1`，可从 Workspace Home 或 workflow definitions 派生本地草案并进入 Draft Designer。
- 当前已接入 `Workflow Draft Designer Editing Model v2`，可在本地新增节点、移动节点、删除非受保护节点，并让 validation inspector、execution plan preview 和 runtime readiness inspector 消费 active draft。
- 当前已接入 `Workflow Draft Node Attribute Editing Model v1`，可编辑 provider / profile、tool ref、RAG ref、input / output contract fields、output mapping 和节点输入输出摘要，并保存 / 恢复到 dev-only saved draft。
- 当前不做 durable persistence，不持久化 validation result / execution plan / runtime readiness，也不发布或执行 workflow。

## 页面状态模型

后续接入 saved draft consumer 时，页面必须至少区分：

- `sample`：离线样例，只用于审查和演示。
- `local_edit`：用户当前本地改动，尚未写入 dev store；这是页面交互状态，不是 production persistence 状态。
- `unsaved_local`：saved draft consumer 显示用户当前本地改动尚未写入 dev store。
- `saving`：正在通过 dev-only consumer 保存。
- `saved_dev_record`：已保存到 dev-only memory store，可读取和校验。
- `version_conflict`：保存时发现当前 saved draft 版本已变化，页面展示 current version metadata，并保留用户当前本地草案。
- `save_failed`：保存失败，必须展示 failure code。
- `read_failed`：读取失败，不能回退成 sample 伪装成功。

页面不得把 `sample` 或 `saved_dev_record` 写成 production persistence，也不得把 `valid_for_review` 写成 publish / run ready。

## 信息架构

页面应保持以下信息分组：

- draft identity：`draft_id`、workspace、application、schema version、draft version 和 saved state。
- editable graph：节点、边、输入契约、输出契约和引用 summary。
- structure editing：本地节点新增、删除保护、节点重排、边重建和 audit edge 保留。
- node attribute editing：节点属性编辑、provider / tool / RAG refs、contract fields 和 output mapping。
- validation：结构校验、契约校验、capability 校验和 risk finding。
- blocked capabilities：executor、confirmation decision、writeback、replay、code、sandbox、agent loop 等禁止项。
- audit metadata：request id、audit ref、actor ref 和 source mode。

## Consumer 接线要求

- 保存前必须从页面状态生成 sanitized payload，而不是提交 UI-only 字段。
- 保存失败必须保留当前本地草案，不得用 sample 或旧 saved record 覆盖用户当前状态。
- 读取失败必须 fail closed，不得静默回退 sample。
- 版本冲突必须展示当前版本 metadata，并要求用户显式选择后续处理；不得自动覆盖本地草案或用 sample 回退。

## 验收方式

- Web build 通过。
- 页面或 consumer 测试覆盖 sample / unsaved / saved / failed 状态区分。
- 相关 Go route tests 通过后，页面才能声明已接 dev-only consumer。
- `./scripts/check-repo.sh --fast` 通过。

## 停止线

- 不实现完整拖拽 builder、分支 DSL 编辑器、publish、run、executor、confirmation decision、writeback、replay 或 resume。
- 不持久化 execution plan、runtime readiness、scenario、review handoff 或 run result。
- 不接真实数据库、Radish OIDC、repository adapter、schema migration、store selector 或 public production API。
