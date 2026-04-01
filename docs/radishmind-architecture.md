# RadishMind 系统架构草案

更新时间：2026-03-30

## 架构目标

`RadishMind` 的目标架构建议冻结为六层：

1. Client Adapters & Context Packers
2. Copilot Gateway / Task Router
3. Retrieval & Tool Layer
4. Model Runtime Layer
5. Rule Validation & Response Builder
6. Data / Evaluation / Training Pipeline

这个拆分的核心目的，是让不同上层项目通过统一协议接入，同时把“项目语义”“工具编排”“模型推理”“安全校验”和“评测闭环”解耦。

## 总体工作流

当前建议按以下顺序理解一次完整请求：

```text
上层项目状态 / 文档 / 附件 / 图像
        ↓
Adapter 打包为统一 CopilotRequest
        ↓
Copilot Gateway 识别任务与项目
        ↓
Retrieval / Tools 获取补充证据并压缩上下文
        ↓
LLM / VLM 生成解释、问题与候选动作
        ↓
Rule Validation 校验风险、目标和结构
        ↓
Response Builder 输出可消费 JSON
        ↓
Adapter 映射回各自 UI / 日志 / 候选提案
```

## 分层说明

### 1. Client Adapters & Context Packers

面向上层项目的集成适配层。

当前建议至少预留：

- `adapter-radishflow`
- `adapter-radish`

职责：

- 将业务上下文转换为统一 `CopilotRequest`
- 对敏感字段做裁剪、脱敏和禁止透传
- 处理认证、会话、超时、重试、缓存和本地回填
- 将结构化响应映射回各自项目的 UI、日志或候选编辑提案

项目重点：

- `RadishFlow`
  - 优先走状态优先打包：`FlowsheetDocument`、`document_revision`、`SelectionState`、`DiagnosticSummary`、`SolveSessionState`、`SolveSnapshot`
  - 控制面相关只打包 entitlement / manifest / lease / sync 的摘要，不透传 token 或 credential
- `Radish`
  - 优先走知识优先打包：固定文档、在线文档、论坛/文档 Markdown、Console 权限知识、附件引用
  - 角色和权限只传最小必要摘要，不透传原始 token、cookie 或安全凭据

不负责：

- 训练模型
- 持有业务真相源
- 直接执行未经确认的破坏性修改

### 2. Copilot Gateway / Task Router

作为 `RadishMind` 的统一入口服务。

职责：

- 接收多模态请求
- 识别 `project`、`task`、`schema_version`
- 路由任务到对应的 retrieval / tool / model 组合
- 统一输出结构化响应
- 提供鉴权、审计、请求追踪与版本标记

建议接口风格：

- HTTP JSON 为主
- 图片和大文件通过对象引用或附件引用传递
- 输出保持统一骨架，但允许项目专属上下文字段

### 3. Retrieval & Tool Layer

这里承载“模型之外的确定性能力”。

职责：

- 文档检索
- 项目语义转换
- 附件 / Markdown / JSON 解析
- 候选动作构建
- 调用外部工具或项目专属只读解析器

当前建议至少拆出两类工具：

- 证据获取工具
  - 读取固定文档、在线文档、论坛正文、附件摘要、结构化状态
- 候选动作工具
  - 在不直接写入业务真相源的前提下，生成可确认的候选编辑或操作提案

原则：

- 能规则化的逻辑，不强行让模型背
- 能提前压缩的上下文，不把全部原始状态直接丢给模型
- 任何高风险动作都只能输出候选动作，不直接执行

### 4. Model Runtime Layer

这里承载真正的 LLM / VLM 推理。

建议分为两类：

- Teacher Models
  - 用于强能力推理、数据蒸馏、标注参考和 PoC 对照
- Student Models
  - 用于本地化、小成本部署和项目内推理实验

当前判断：

- `RadishFlow` 第一阶段不应把全部任务都压成截图推理；结构化状态和诊断解释优先
- `Radish` 第一阶段以文档、内容和 Console 知识问答为主，VLM 只在附件或截图理解场景补充
- `minimind-v` 当前更适合作为 student/base 候选和实验底座，而不是预设主模型

### 5. Rule Validation & Response Builder

这里承载模型输出后的硬约束和结构收口。

职责：

- 校验响应结构、字段完整性和版本兼容
- 检查目标对象、风险等级和确认要求
- 过滤不允许的动作或越权建议
- 生成统一引用、证据和 `requires_confirmation`

当前原则：

- 模型输出默认是建议，不直接成为最终状态
- 高风险动作必须带 `requires_confirmation`
- 若模型证据不足，应允许退化为检索或模板式回答

### 6. Data / Evaluation / Training Pipeline

职责：

- 构建训练样本
- 管理合成数据与人工校正数据
- 执行离线评测
- 跟踪模型版本、任务表现和回归

第一阶段优先建立：

- `RadishFlow`
  - 基于 `FlowsheetDocument`、选择集、诊断摘要、求解快照和控制面错误摘要的合成样本
  - 后续再补基于真实画布的截图样本
- `Radish`
  - 基于固定文档、在线文档、论坛 Markdown、Console 文档和附件引用的样本
- 共享指标
  - 结构合法率
  - 引用命中率
  - 建议可执行率
  - 风险分级正确率
  - 任务级回归稳定性

## 推荐仓库结构

建议先采用如下结构：

```text
RadishMind/
├─ README.md
├─ docs/
├─ contracts/
├─ services/
│  ├─ gateway/
│  ├─ orchestrator/
│  ├─ tools/
│  └─ evaluation/
├─ adapters/
│  ├─ radishflow/
│  └─ radish/
├─ datasets/
│  ├─ synthetic/
│  ├─ annotated/
│  └─ eval/
├─ training/
│  ├─ configs/
│  ├─ scripts/
│  └─ notebooks/
├─ prompts/
│  ├─ system/
│  ├─ tasks/
│  └─ eval/
└─ experiments/
```

## 关键设计口径

- 统一先走结构化协议，不让不同项目各自发散
- 协议采用“通用骨架 + 项目专属上下文块”，而不是强行做一个业务超集
- 当前主实现栈收口为 `Python`，优先统一评测、数据处理、模型适配和自动化校验工具链
- `RadishFlow` 优先走状态优先上下文，截图是补充，不是第一阶段唯一中心
- `Radish` 优先走知识优先上下文，重点是 Docs / Wiki / Forum / Console 语义
- 模型输出默认是建议，不直接成为最终状态
- 训练数据优先从自有项目生成，不依赖大量外部脏数据
- 评测必须从第一阶段就开始建立
- 部署形态允许本地、局域网和远端三种模式并存，但当前不先锁定最终部署形态
