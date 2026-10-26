# RadishMind 产品机会池

更新时间：2026-10-26

## 文档目的

本文档记录 `RadishMind` 长期产品机会，避免把临时讨论散落在对话里。它不是路线图，不代表功能已经承诺或实现；进入排期前先明确真实任务、现有缺口、收益与验证方式；只有必要的高风险边界才拆实施任务卡。

当前主线为[本机持久工作台交付](features/user-workspace/local-persistent-workbench-delivery-v1.md)，后续选择以[路线图](radishmind-roadmap.md#后续开发方向2026-10-26)为准。本文候选不自动进入开发，也不用于填补真实模型试用暂停期间的等待。

当前正式产品面仍以 [产品范围](radishmind-product-scope.md) 为准：`User Workspace`、`Admin Control Plane`、`Model Gateway / API Distribution`、`Workflow / Agent Runtime`。

## 产品增强候选（未排期）

### 1. `Model API Routing Studio`

给管理员可视化配置模型 API 分发策略：谁能用哪些 provider/profile/model、额度多少、成本上限多少、是否允许 fallback、敏感任务禁用哪些 provider。

- 价值：直接连接当前 provider runtime、API distribution、quota、cost 和 policy。
- 依赖：API key 生命周期、quota policy、provider health signal、secret ref readiness、audit record。
- 停止线：不在没有租户 / API key / audit 基础前做复杂路由 UI。

### 2. `Workflow Marketplace`

用户或团队可以发布 workflow、prompt app、RAG app 和 agent template，管理员可以审核、上架、禁用、限制模型和额度。

- 价值：让 User Workspace 从“自己搭流程”扩展成团队复用和分发。
- 依赖：workflow definition、版本、权限、运行记录、审核状态。
- 停止线：不先做公开市场；首版只考虑 Radish 体系内部团队空间。

截至 2026-10-26，内部团队空间的首版已收敛为[工作区 Workflow 模板目录、审查与受控派生（开发 / 测试态）v1](features/workflow/workspace-workflow-template-catalog-review-controlled-derivation-dev-test-v1.md)，既有规范契约、三种存储、Web、审查 / 上架 / 派生与双数据库验收已完成关闭，不续批。公开 Marketplace、跨 workspace / tenant、评分推荐、收费和第三方分发仍留在机会池，不随首版打开。

### 3. `Eval-as-a-Feature`

候选方向是让重复版本迭代更容易保留评测材料和人工审查证据。已有 Plan / Campaign、同版本配对、Case / Suite 与受限 Schedule；[跨版本审查](features/user-workspace/application-cross-version-regression-review-dev-test-v1.md)暂不排期。现有运行元数据比较不判断答案质量，自动触发或模型裁判须由真实使用需求另行决定。

- 价值：把本仓库已有 evaluation / governance 能力变成用户可见产品能力。
- 依赖：eval sample schema、run record、应用版本、发布门禁、人工复核记录。
- 停止线：不把机器指标直接写成生产可发布；高风险应用必须保留人工复核。

### 4. `AI Cost Ledger`

按租户、用户、应用、workflow 节点、provider、model、失败重试、缓存命中记录成本账本。

- 价值：让管理员知道钱花在哪里，也给路由策略和 quota 提供真实依据。
- 依赖：request trace、token / cost estimator、provider pricing profile、workflow run record。
- 停止线：不在 provider pricing 和 trace 未稳定前做账单承诺。

### 5. `Policy Sandbox`

管理员定义策略后，用户构建 workflow 时即时看到策略阻断原因，例如外部 HTTP tool 禁用、敏感数据禁止某些 provider、特定 tool 必须人工确认、输出必须脱敏。

- 价值：把安全治理前置到构建过程，而不是等运行后报错。
- 依赖：policy schema、tool registry、risk classification、tenant / role binding、blocked action contract。
- 停止线：不把 sandbox 结果写成真实执行授权；最终执行仍要走 runtime policy。

## 后续候选

### 6. `Provider Reliability Score`

持续记录 provider/profile 的错误率、延迟、timeout、schema failure、stream failure 和成本波动，生成可靠性分数，供路由策略和运维看板引用。

### 7. `Prompt / Workflow GitOps`

prompt、workflow、provider route 和 quota policy 可导出为 JSON / YAML，走版本 diff、review、回滚和环境 promotion。

### 8. `Copilot Handoff Protocol`

为 `RadishFlow`、`Radish` 和后续项目固定结构化 handoff：建议动作、风险、证据、确认要求和回滚提示，让 AI 只交付可审计候选动作。

### 9. `Workflow Replay / Time Travel`

候选目标是从精确历史版本与保留证据解释回答变化。历史重读与重新执行必须分开；输入 / 正文保留、外部副作用、授权和费用边界未确定前，不开放历史重放或恢复执行。

### 10. `Personal AI Toolbox With Team Governance`

用户可以搭个人小工具，团队管理员统一管理模型、额度、tool 权限和审计，实现个人效率和团队治理的平衡。

## 推进原则

- 当前按本机交付、真实效果和实际共享 / 接入需求选择工作，不再沿用早期 Gateway 优先的建设顺序。
- 候选进入路线图前在功能专题明确目标用户、现有阻塞、范围、验收和未完成项；不默认逐项创建任务卡。
- 不把本机会池写成 roadmap commitment。
- 不把当前本地 ops console 扩写成正式用户端、生产管理端或 marketplace。
- 不为产品机会新增后端语言栈；默认继续使用 Go control plane / gateway、Python model / eval / worker、TypeScript/Vite frontend。
