# 受控开发阶段 0：代表任务、SDK 与执行合同

更新时间：2026-11-11

状态：离线准备、人工任务复验、SDK LLM 子合同及单尝试持久事务 / 当前资格 / 内部通道固定响应验证已完成；受限 SDK Agent / 完整允许事件、工具往返及等待取消已通过固定响应验收；固定 Go 沙箱 / 受限工具 / 归属终止已真实验收；开发产物合同与 SQLite 显式保存 / 导出已实现并定向验证；真实 SDK 联合技术任务已通过；采用复评已完成但正式采用暂不通过：原生 HTTP 已通过固定上游技术验证，阶段 1 合同范围与 A 级画板已确认，许可交付仍待闭合。

所属功能：[受控软件开发团队](../features/workflow/controlled-software-development-team-v1.md)。本卡集中记录阶段 0 探针、单尝试控制合同与工具边界，不为各层再建任务卡或检查器。

## 2026-11-10 规划收口接续

本卡继续记录原有技术结果。单 Agent [公开合同](../contracts/development-task-delivery-v1.md)及六幅代表面的修订要求已确认。原画板修订出图已获项目所有者最终视觉确认。许可按具名来源缺口裁决并完成一个最终候选的交付验收。具体退出条件和阶段 1 后的真实单 Agent 价值复核见[实施专题](../features/workflow/controlled-software-development-team/implementation-and-evaluation-plan-v1.md)。

11 月 11 日完成 Pillow 单 wheel 候选首轮六次输入取证：五份 blob 精确匹配，multibuild 快照按 Git 类型 / 可执行位、路径、长度和内容摘要完整核对，见[结果](../../scripts/eval/development_probe/sdk_variant/pillow-fixed-inputs-result-20261111.json)与[周志](../devlogs/2026-W46.md#2026-11-11-pillow-首轮取证与交付阻塞复核)。本窗口已结束；构建镜像 / 工具、嵌套源码和最终交付未完成，保留单一候选，阶段 0 不退出。

## 本轮范围

规划归属：Workflow / Agent Runtime → 受控软件开发团队 → [实施顺序与效果验证](../features/workflow/controlled-software-development-team/implementation-and-evaluation-plan-v1.md)。本卡是阶段 0 任务结果源；框架、运行状态、隔离与界面职责子专题提供关联合同。下表完成项只证明对应技术范围，阶段退出仍须满足实施专题的四项条件。

项目所有者要求按建议继续推进：先落实 SDK / Gateway 兼容验证与代表任务，再推进单 Agent 隔离交付。现阶段已实现离线计划 / 工具请求校验、副本准备及独立 SDK LLM 契约测试；SDK 候选、执行镜像、正式 Gateway 扩展及代码沙箱不能凭准备结果视为已接入。

| 工作 | 当前结果 | 完成依据 / 剩余项 |
| --- | --- | --- |
| 固定代表任务 | 完成 | 标准库 Go 列表接口，三份文件及摘要固定，预置 8 项测试；只可修改 `list.go` |
| canonical 探针合同 | 完成 | 两份严格 schema、Go 消费者；拒绝缺失 / 未知 / 重复 / 大小写别名字段、空值、尾随 JSON、越界路径和命令扩权 |
| 副本与候选身份 | 完成 | CLI `check / prepare / inspect`；新目录拒绝覆盖，检查保护文件、文件集合及摘要，未开放执行入口 |
| 合成任务复验 | 完成，人工方式 | 基准两项预期失败；人工修正候选后 8 项通过；基准未变，不是 SDK 生成或沙箱执行 |
| SDK 候选依赖窗口 | 已完成并关闭 | 1.59.0 / arm64 / Python 3.12.16，138 个依赖完整 wheel 哈希锁及按锁复建通过；镜像与 API / 依赖元数据已保留，元数据复评发现必需专有代理依赖，许可路线未闭合 |
| SDK LLM 子合同 | 7 项通过 | 固定工具 / 观察值往返两次请求；非法输入、未知工具、普通文本误判、HTTP 错误 / 超时与 usage 缺失观察；不运行 executor 或完整 Agent |
| 单尝试控制合同 | 已完成并扩展 | 七份计划 / 控制 / 工具 schema；严格消费、SDK summary 显式分离、可信提议 / 预留摘要、并发预留、未知用量、取消 / 迟到核账离线验证；不是认证或执行服务 |
| 持久尝试与内部通道 | 本轮完成 | 共享 SQLite 独立组件；提议消费 / 预留 / 事件事务，当前身份 / 应用 / 模型 / 定价读取，已验证 Session 的可信 Go 签发，独立回环监听器及固定原生往返；默认关闭，工具只入队 |
| 受限 SDK 工作进程 | 固定响应验收通过 | 真实 Agent / LocalConversation 经 Go 专属通道完成工具往返、允许事件、错误不重试、未知用量与模型等待取消；父存活管道失联退出已验，工具观察值仍为合成值 |
| 固定 Go 沙箱与工具 | 真实容器验收通过 | 私有副本读取 / 原子替换，按候选摘要测试，只读源码、资源限制、硬期限、取消 / 遗留归属清理与输出超限；SDK 收尾接线已定向验证 |
| 开发产物保存 / 导出 | 本轮实现并定向验证 | 独立负责模块、严格包合同、计划 / 基准 / 候选 / patch / 测试绑定、显式确认、独立权限、SQLite 配额 / 幂等 / 重启与失败恢复；未开放 HTTP / Web |
| 联合任务 / 正式采用 | 联合技术验收通过，复评暂不通过正式采用 | 同次任务覆盖 SDK 修改候选、真实代码测试、保存 / 导出、旧测试失效与命令中取消；资格 / 提供方为合成适配，旧 Gateway 保持建议语义；原生 HTTP 已于 11 月 5 日通过固定上游技术验证；新候选已移除默认专有代理链并复验，阶段 1 公开合同与最终视觉已确认，剩余许可交付 |

本轮阶段 0 设计：公开操作 / 权限、SQLite 原子启动、独立执行槽、请求回执、取消与重启核对、保存 / 导出及保留边界已写入公开交付合同；六个 Pencil 代表面已保存并静态复核。具体合同与最终视觉已于 11 月 10 日获得项目所有者确认，公开 API / Web 仍未实现。

## 已确认的依赖调整

项目所有者已授权并完成 `1.59.0+radishmind.1` 遥测可选化候选的构建和复验。[补丁、来源与运行记录](../../scripts/eval/development_probe/sdk_variant/README.md)绑定独立 119 项锁和实际镜像；4 项遥测检查、7 项 LLM 子合同、完整 SDK 工作进程及联合正常 / 旧测试失效 / 命令中取消通过。默认集合不含 Laminar / 专有代理，余下依赖版本和摘要不变；SDK wheel 的 MIT 正文与修改记录已核对。原始锁 / 镜像保持历史对照，新版本绑定新计划摘要；剩余依赖 / 基础镜像许可未全部闭合，不登记正式采用。本次所属容器已清理，运行窗口结束。

## 合同与边界

- `contracts/development-task-probe-plan.schema.json` 只允许 `fixed_response_only` 合成任务；限制模型 / 工具次数、请求 / 输出大小、命令与总时限。时间总预算必须不小于命令上限。
- `contracts/development-task-probe-tool-request.schema.json` 只描述 `read_file / replace_file / run_tests`。所有字段必填；替换绑定旧文件摘要，测试绑定候选摘要，不接受 shell 或自选命令。Go 同时检查计划身份、序号上界与替换内容的 UTF-8 字节大小。
- 旧探针的工具请求解析**不执行工具**，自身不实现单调序号、重放拒绝、当前摘要核验或预算预留；这些必须在执行控制层落实后才能接收模型请求。
- `workspace.json` 与候选摘要是准备材料，不是正式交付包或测试通过声明。树摘要为计划顺序下 `path + NUL + sha256 + LF` 的 UTF-8 SHA-256；计划摘要为 Go 固定结构紧凑 JSON 的 SHA-256，不是任意 JSON 文本摘要。
- 0400 权限、文件检查和人工宿主机测试不构成不可信代码沙箱。后续代码容器不得持有凭据、Docker socket、原工作树或平台数据；测试文件需要操作系统只读边界。
- 现有 Agent / Copilot 的 `allow_tool_calls=false`、30 秒调用与 Gateway 协议保持。这里的离线 schema 不是正式 HTTP 授权、生产执行资格或团队资源协议。

## 验证与下一工作

已完成 Go race / 覆盖率定向测试、schema / 冻结摘要 / 锁生成测试，以及合成任务先失败后通过的人工复验。完整命令、摘要及仓库门禁记录见 [W44 周志](../devlogs/2026-W44.md#2026-10-31-受控开发代表任务与离线探针)。

已按明确授权完成候选镜像、完整锁、接口核对与固定响应子合同检查，详情见[SDK 检查说明](../../scripts/eval/development_probe/README.md)和 [W44 运行记录](../devlogs/2026-W44.md#2026-10-31-sdk-候选镜像与固定响应契约验收)。上述合同、持久尝试 / 提议 / 事件事务、真实资格与专属回环接线已实现，见[尝试控制合同](../contracts/development-attempt-control-v1.md)。随后已按单独授权复用锁定镜像，以真实 SDK Agent 验证受限工作进程、完整允许事件、两次模型请求 / 合成工具观察值往返、模型失败不重试、工具失败、缺失用量、等待取消及父存活管道失联退出，见[本轮记录](../devlogs/2026-W44.parts/development-sdk-worker-20261031.md)。适配显式关闭 SDK 默认重试 / 动态上下文等旁路，`summary` 单独转换，Go 保留未知账目，完成须等待真实 OS 退出和空进程组。固定 Go 探针代码沙箱、受限文件 / 固定测试工具和资源归属终止已实施并真实验收，详见[沙箱合同](../contracts/development-code-sandbox-v1.md)及[运行记录](../devlogs/2026-W44.parts/development-code-sandbox-20261031.md)。开发产物保存 / 导出合同和内部 SQLite 链已完成，见[开发产物合同](../contracts/development-artifact-v1.md)及[本轮记录](../devlogs/2026-W44.parts/development-artifacts-20261031.md)。真实 SDK 与代码容器同次任务的修改、测试、保存 / 导出及取消已按独立授权完成，见[联合验收记录](../devlogs/2026-W44.parts/development-joint-20261031.md)。[采用复评已完成](../devlogs/2026-W44.parts/development-adoption-20261031.md)，暂不通过正式采用；遥测可选化候选已构建并独立复验；11 月 5 日已核对 OpenSSL 精确归档与 CNRI 最终 1.6 正文、接通五包告知的离线准备与构建配方（未重建镜像），并完成原生 HTTP 配置 / 凭据接线、两轮工具往返及固定上游正反例验证，见 [W45 周志](../devlogs/2026-W45.md)。原生组件离线对应已覆盖 26 个 wheel / 63 个 ELF / 10 份 SBOM，八个外部库的实际镜像文件 / 归属已静态核对，容器未启动且已清理；随后已核对 Pillow 静态符号 / 来源差异及 hf-xet 的两项 MPL 来源，六包 14 份 wheel 告知与基础镜像 pip 21 份内嵌告知已接入离线准备及构建校验（未重建镜像）；覆盖源码采用伴随精确镜像的候选布局，范围已扩至 119 个 wheel / 61 个 Debian 源包 / 10 个 Alpine origin，Pillow 来源差异、certifi 生成步骤及 pip / Go 内嵌源码对应已核对；历史构建输入和实际源码交付仍待完成。11 月 5 日本轮又完成双镜像工具链静态核对与有界历史来源取证，两个停止态容器已删除，下载窗口已关闭；Go 853 个工具链内嵌文件与精确源码一致，CPython 78 个扩展已静态枚举。正式采用仍有历史输入、内嵌告知和最终交付缺口。固定 Go 单 Agent 的[公开交付合同](../contracts/development-task-delivery-v1.md)及六个 Pencil 代表面已形成可评审设计，具体合同与画板修订要求已获项目所有者确认，源画板已修订和出图，最终视觉已确认。剩余缺口依据[11 月 5 日采用评审](../features/workflow/controlled-software-development-team/framework-and-gateway-integration-v1.md#2026-11-05-框架采用评审)，当前顺位以[收口入口](../radishmind-current-focus.md#next-actions)为准，不将更多元数据自动等同采用通过。四项剩余退出条件见[实施专题](../features/workflow/controlled-software-development-team/implementation-and-evaluation-plan-v1.md#阶段-0-采用复评与剩余退出条件)。阶段 0 尚未退出；暂不真实用户试用或真实模型，不启动固定 A/B/C/D/R 团队，不自动合并 / 推送 / 发布。
