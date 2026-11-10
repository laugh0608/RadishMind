# 阶段 0 采用复评记录（2026-10-31）

起点为干净 `dev@095ed28d`。项目所有者要求核对 SDK 许可证、提供方原生传输、阶段 1 UI / 实现范围与剩余退出条件。本轮只检查证据和代码、更新设计文档；未安装或更新依赖，未启动 Docker / 服务 / SDK，未调用真实模型，未开展用户试用。

## 结论

**OpenHands SDK 保留首选候选，正式采用暂不通过；阶段 0 不退出。** 联合技术链已通过，不再重复列作未完成。剩余四项为许可路线、原生 HTTP 适配证据、阶段 1 合同 / 范围确认及 A 级视觉依据，具体进入条件以[实施专题](../../features/workflow/controlled-software-development-team/implementation-and-evaluation-plan-v1.md#阶段-0-采用复评与剩余退出条件)为准。本轮没有替项目所有者批准新的依赖方案、公开 API、实现范围或视觉设计。

## 许可证证据

对照提交锁与已保存安装报告，核对全部 138 项的规范化包名、版本及 wheel SHA-256，一致。安装元数据共 139 项，多出的 `pip==25.0.1` 属于基础镜像；它不是锁中的新增依赖。完整元数据清单位于忽略目录 `tmp/development-adoption-20261031/license-inventory.json`，仅登记声明，不把元数据等同完整许可审查。

- `locked-installation.json`：`sha256:73f37436acbb905935b083cdee853be5c044db8f8beaba0b937fcc00ac261e94`。
- `sdk-inspect.json`：`sha256:24f04a0f1743e78de19b153a1ae31566a65feac36ff5971a1c1dc45411ea6634`。
- `license-inventory.json`：`sha256:2d5bc6abdc5283efef299bc99bc9a0c4894555f9e5e73f050ca6c6c9b1f9871a`。

| 精确分发包 | 发现 | 处理状态 |
| --- | --- | --- |
| `openhands-sdk==1.59.0` | 安装元数据无许可字段；锁定 wheel 未找到 LICENSE / NOTICE / COPYING 文件；[上游精确标签](https://github.com/OpenHands/software-agent-sdk/blob/v1.59.0/LICENSE)为 MIT | 补正式分发告知与来源证据，不推定镜像已含许可证 |
| `lmnr==0.7.67` / `lmnr-claude-code-proxy==0.2.0` | SDK 必需 `lmnr>=0.7.60,<0.8.0`，后者必需 `lmnr-claude-code-proxy>=0.1.22`，没有 extra 条件；代理的[官方元数据](https://pypi.org/project/lmnr-claude-code-proxy/0.2.0/)为 `LicenseRef-Proprietary` | 当前采用阻塞；关闭遥测不移除该分发包，不能凭未调用代理推定可分发 |
| `agent-client-protocol==0.12.1` | 元数据无声明；锁定 wheel 的 `licenses/LICENSE` 含 Apache-2.0 正文 | 元数据空缺已补证，仍需保留发行物告知 |
| `certifi==2026.10.23` | 元数据为 MPL-2.0 | 核对精确正文与交付义务 |
| `orjson==3.13.0` | 元数据为 `MPL-2.0 AND (Apache-2.0 OR MIT)` | 不因含 MIT 选项忽略 AND 的 MPL 条款 |
| `tqdm==4.70.1` / `regex==2026.9.29` | 前者 `MPL-2.0 AND MIT`，后者复合 Apache-2.0 / CNRI-Python 声明 | 核对各部分适用正文 |
| 其他元数据不充分项 | `fastuuid` / `Jinja2` 只给 BSD 分类，`lupa` 为 MIT style，`pyperclip` / `wrapt` 为笼统 BSD，`python-dateutil` 为 Dual License，`socksio` 的 UNKNOWN 与 MIT 分类并存；另有仅分类声明项 | 不批量猜测 SPDX 细项，正式采用前核对正文 |

通过冻结报告的三个精确 wheel URL，只读内存中的 ZIP 并验证完整 SHA-256，提取许可证文本作为证据；没有安装、执行或保存可执行 wheel 内容。记录在同目录的三个 `*-license-evidence.json`。这次静态读取仅覆盖 SDK、代理与 ACP，未完成全部原生捆绑库、基础 Python / OS 镜像及联合测试 Docker CLI 的许可核验，也不声明全部依赖开源或无其他许可义务。

- 代理 wheel：`sha256:7efa6d4d6a19b898a9b8d538fee93495cab4e62ec30c895a6ca0ad1267420f47`；`LICENSE.md`：`sha256:c188d21a65223f2a4786d2fefb8620096d4f9d950dede63ffe9b7a712c6d954f`。
- ACP wheel：`sha256:aaf3cc301ed87d9e2c4ccb91adfe1cd58784355f6ce7e513f50e7ca2391862ef`；正文：`sha256:cfc7749b96f63bd31c3c42b5c471bf756814053e847c10f3eb003417bc523d30`。
- SDK wheel：`sha256:ecef4942b9294a4b6b6f62d3825b8738d69772dd4c2e2754b73b4c1a8a90012f`。

代理的许可证正文将使用关联至 [Laminar 服务条款](https://laminar.sh/policies/terms)。2026-10-31 查看页面的更新时间为 2026-08-11；条款区分个人 / 内部用途与再分发，后者需要额外依据。现有材料不足以批准将该组件随 RadishMind 镜像再分发；这是本项目采用判定，未作“所有使用违法”的推论。建议先核实许可和拟交付方式，或提出经过批准的依赖替换路线；本轮不联系上游，不接受条款，不删必需依赖，不重锁或降版。

## 原生传输与阶段边界

代码核对：`developmentcontrol/runtime_contracts.go` 定义 `NativeBackend.Policy / Complete`，`httpapi/development_channel.go` 仅接受注入。实现都位于 `runtime_test.go`、`development_channel_test.go` 和 `joint_fixture_test.go` 等测试文件，尚无非测试的提供方 HTTP 后端。`development_qualification.go` 对当前身份、应用、模型档案、凭据配置和定价的校验不证明工具协议兼容；旧 runtime 的建议响应解析也不能替代原生工具传输。

建议先落实单档案非流式 Go 适配，以真实 HTTP 固定上游验收；覆盖实际配置 / 凭据接线、请求工具身份、错误、用量未知、取消、禁止重试 / 重定向及响应上限。具体范围已写入[框架专题](../../features/workflow/controlled-software-development-team/framework-and-gateway-integration-v1.md#原生传输的建议实施范围待确认)，尚未实施。真实厂商 / 模型连通性与质量另验，暂停状态保持。

阶段 1 建议固定 Go 单 Agent 纵向交付：User Workspace 内确认计划、发起 / 取消、事件、候选 diff / 测试、保存 / 导出与历史重读；需要公共任务合同 / 权限、生命周期接线与重启核对。内部 Go 接线不是公开 API，阶段 0 的管理员资格也不能自动成为正式任务权限。通用 Git 来源、依赖安装、固定团队、群聊和自动合入不混入该首版；如需通用来源，应先明确独立合同。

[UI 复评](../../features/workflow/controlled-software-development-team/collaboration-workspace-ui-v1.md#阶段-1-单-agent-范围与设计判定)评分 `1 / 2 / 2 / 2 / 1 = 8`，属于 A / 完整 Pencil。需代表面及不可推导风险 / 窄屏状态的视觉依据，尚无画板、视觉批准、React 或浏览器证据。现有依赖可支持只读文本 diff，不默认安装编辑器。

## 验证与清理

本轮未修改代码、schema、依赖锁、镜像或执行配置。许可证清单与安装报告 / 提交锁逐项一致性核对通过。`./scripts/check-repo.sh` 全量通过（退出码 0，日志 `tmp/development-adoption-20261031/check-repo-full.log`），未跳过平台 Go 测试。Docker / SDK / 联合容器测试入口环境变量均未设置，未执行这些需独立授权的运行试验。收尾文字调整后，文本、篇幅、中文规范、Markdown 链接及 `git diff --check` 复核通过；仅保留仓库既有篇幅提示，本周周志仍在 350 行提示线内。

保留忽略目录中的元数据 / 许可证文本证据及检查日志，无本轮服务、容器或后台进程需要清理。提交只含专题、规划、任务卡与周志，不含临时证据、运行代码或许可授权文件；没有推送或发布。
