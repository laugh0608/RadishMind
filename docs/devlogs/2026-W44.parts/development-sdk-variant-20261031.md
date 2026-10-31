# SDK 遥测可选化候选准备与复验（2026-10-31）

起点为干净 `dev@32990425`。项目所有者在许可证路线核对后明确要求按推荐继续，选择保留 SDK 1.59.0 基线、将 Laminar 改为可选依赖；准备阶段只落实可审阅候选，不沿用此前已结束的 Docker / 依赖运行窗口；后续单独授权的构建与复验见下节。

## 准备阶段已完成

- 固定 PyPI `openhands_sdk-1.59.0.tar.gz` 的 URL 与 SHA-256；上游精确文件与 MIT 正文通过 Git blob 和本地 SHA-256 双重核对。补丁派生 `1.59.0+radishmind.1`，只改版本 / 许可证元数据、Laminar 可选依赖和缺失依赖错误；关闭遥测的执行行为沿用原代码。
- `source.json` 绑定前后文件、补丁与来源。MIT 正文和修改记录进入新 wheel 的许可文件声明；没有修改代理代码或为它更换许可证。
- 离线源码准备器拒绝摘要不符、路径逃逸、链接、特殊成员和已有目标；只写仓库临时新目录，失败留痕。独立 Docker 配方先正常解析，再通过完整 wheel 哈希锁离线安装和 `pip check`，禁止候选默认闭包包含 `lmnr` / 代理包。
- 构建工具单独固定 `setuptools==83.0.0 / wheel==0.46.3 / packaging==26.3`，只准备哈希锁，未安装。运行依赖用原版本约束正常求解，尚未生成新锁或推测新依赖数。
- canonical 探针 schema 与 Go 消费者只新增精确变体，计划摘要随版本变化；Python 工作进程拒绝其他变体，对当前变体拒绝已安装 Laminar 分发包。原始计划 / 138 项锁保留，不混淆旧证据。
- 联合入口支持显式本地 SDK 镜像 digest，仍禁止 tag / 自动拉取，配置检查与清理保持。已在隔离目录准备变体计划和三文件输入，使用已有模块缓存交叉编译检查程序；复用原 Linux Docker CLI，SHA-256 为 `37a53fed3445028a49443fd6f9338ccedeeeda9776de5d7fd4d7e7c936b789cc`。

补丁、离线入口、完整构建 / 运行命令、资源边界与清理要求集中在[候选说明](../../../scripts/eval/development_probe/sdk_variant/README.md)，不另建任务卡或检查器。

## 准备阶段验证与未完成项

上游两个精确源码文件的补丁检查、实际应用及结果哈希核对通过，没有导入或执行 SDK。24 项开发域 Python 离线测试通过，包含源码包边界、候选 wheel 路径 / 摘要错绑、必需代理拒绝、版本与计划身份；Go 探针、尝试控制、沙箱定向 race 通过。首次本地 wheel 测试发现 macOS 临时路径经 `/private` 规范化后不一致，已改为比较规范化路径并支持 URL 百分号编码，负例继续拒绝不同来源与篡改内容。

日志位于 `tmp/development-sdk-variant-20261031/`，目录内还有只读核对用上游文件、补丁结果和待运行 Go 二进制。`./scripts/check-repo.sh` 全量通过（退出码 0，日志 `check-repo-full.log`），Docker / SDK / 联合测试入口显式关闭；新增源码准备正例再次定向通过，文本、链接与 `git diff --check` 收尾复核通过。本轮没有安装依赖、构建镜像、启动容器 / 服务、接入真实提供方或调用真实模型；新候选真实导入、缺失遥测错误、LLM 往返、完整事件 / 取消、同次候选 / 测试 / 保存 / 导出及旧证据失效均未运行，不能继承旧镜像通过结论。其他依赖告知、MPL 对应源码及基础镜像许可仍未整体闭合，正式采用与阶段 0 退出保持未完成。

## 后续独立授权构建与复验

候选准备提交为 `f816bfa9`。项目所有者随后明确批准固定源码 / 依赖下载、候选镜像构建及可信控制容器访问 Docker 管理通道，授权范围沿用候选说明的隔离路径、资源限额和结束清理。没有安装宿主机依赖、启动 Docker 守护进程、发布宿主端口或连接真实模型。

### 构建与发行物核对

- 下载的源码归档为 813527 字节，SHA-256 与 `source.json` 一致；准备器验证源码、补丁和修改后摘要，336 个准备文件再次核对通过。
- 第一次构建因沙盒禁止写入 Docker Buildx 的活动记录而在开始前失败，日志保留为 `resolution-build.log`；在已批准范围内取得对应执行权限后，第二次构建成功，日志为 `resolution-build-02.log`。没有忽略依赖约束或安装后删包。
- 正常解析得到 119 个分发包，原锁为 138 个；`lmnr`、`lmnr-claude-code-proxy` 及其不再使用的依赖共 19 项退出集合，其余依赖版本和 wheel 摘要逐项不变。源码派生 SDK wheel 的 SHA-256 为 `9c46398603f5e6cf39e33525e9731c519bf8b2040dde12915092ee005c0b36d4`；新锁 SHA-256 为 `3e2f6bc2ca091acd36cb201ae5f66ed69b50d080af4b55c32314e0b1283983a3`。
- 最终镜像安装禁用网络，按全部 wheel 哈希安装并通过 `pip check`。119 项实际安装记录与解析报告的版本 / 哈希完全一致；基础镜像另带 `pip`，元数据枚举为 120 项。SDK wheel 内 MIT 正文、修改记录和修改后的 Python 文件均匹配来源清单。
- `linux/arm64` / Python `3.12.16` 镜像 ID 为 `sha256:912f0a65e355629b9578c4c80d2da0b3b0881e467874b476dfdc3e3f816e4acf`，运行均引用此 ID。只主张本次 wheel 冻结安装可核对，不主张重新源码打包或镜像逐位可复现。

独立锁与[验收摘要](../../../scripts/eval/development_probe/sdk_variant/acceptance-20261031.json)已登记；完整解析 / 安装报告、wheel、镜像配置及日志留在隔离目录。旧锁和镜像没有替换或删除。

### 独立运行结果

| 验证层 | 结果与边界 |
| --- | --- |
| 遥测可选依赖 | 4 项通过：精确版本 / MIT 文件、默认无 Laminar / 代理包、SDK 导入与普通观察、四个环境入口及直接根 span 缺失依赖拒绝 |
| SDK LLM 子合同 | 7 项通过：工具往返、非法输入 / 未知工具拒绝、文本不冒充工具、错误不重试、超时与未知用量 |
| 完整工作进程 | 三个主测试及子场景通过：真实 Agent 工具往返、模型 / 工具失败、缺失用量、等待取消和父存活管道断开 |
| 联合正常候选 | 48 个连续事件，8 项固定 Go 测试通过，绑定同一候选摘要；最终 `completed`，导出验证状态为 `command_passed` |
| 联合旧证据失效 | 48 个连续事件，保留原基准两项失败记录，修改后证据状态为 `stale`，不沿用旧候选摘要 |
| 联合命令中取消 | 41 个连续事件，128 毫秒内确认工作进程退出和进程组为空；最终 `outcome_unknown`，导出验证状态为 `unknown` |

三个联合场景均覆盖预览不自动保存、错误确认摘要拒绝、SQLite 写入失败无半成品、同键重试 / 幂等、关闭后重启读取、严格导出重读、撤销导出权限后读取仍独立可用；重启没有重放模型。固定响应请求数分别为 4 / 4 / 3，账本与事件由 Go 核对。

SDK 实际 UID / GID 为 `65532`、附加组为空、capability 全零、`no_new_privileges` 生效；管理 socket 和控制数据访问拒绝全部通过。代码容器保持既有只读源码、无网络与资源约束。四个 SDK 检查容器和一个联合控制容器均正常退出、无 OOM，并已删除；三份资源账本没有遗留代码容器，也没有由宿主补救删除。本轮测试监听器随进程退出，运行窗口已结束，镜像 / 缓存 / 失败日志保留。

本次结果支持登记遥测可选化技术候选；余下依赖告知、MPL 对应源码、基础镜像与工具分发依据仍未整体闭合。原生 HTTP 传输、阶段 1 公开合同 / 实现范围及 A 级视觉依据仍待确认，正式采用与阶段 0 退出保持未完成；未开展真实模型、用户试用、正式 Web 或固定团队。

收口验证：`./scripts/check-repo.sh` 全量通过（退出码 0，日志 `check-repo-acceptance-full.log`）；默认 Docker / SDK / 联合入口关闭，真实运行证据以上述独立窗口为准。提交前再次核对新锁、来源清单和已登记日志 / 导出文件摘要，文本 / 链接与 `git diff --check` 复核通过。

## 剩余许可只读取证与处理方案

起点为干净 `dev@964885f9`。项目所有者要求先核对已有证据、列清剩余依赖及基础镜像许可缺口，再说明新增下载或运行窗口。本次只读本地 wheel / 归档 / ELF / Go 构建信息及镜像元数据，并查阅官方许可说明；未下载依赖或源码包，未安装、构建、创建 / 启动容器或修改依赖锁。

### 已有证据可以支持什么

逐项核对 119 个 wheel 与冻结解析报告的 SHA-256，全部一致。117 个包带许可相关文件，共 141 个文件（含正文、NOTICE、作者或修改告知）；文件存在不等于全部许可义务已经满足。另识别 26 个原生包、63 个 ELF 对象，静态读取动态依赖表，没有导入或执行包。清单保留文件路径、原文与摘要，缺口处理不能仅依赖顶层 SPDX 字段。

- `fastuuid==0.14.0`、`Jinja2==3.1.6`、`pyperclip==1.11.0` 的正文可辨认为三条款 BSD；`socksio==1.0.0` 有明确 MIT 正文，补清此前笼统 / UNKNOWN 声明。`python-dateutil==2.9.0.post0` 正文说明 BSD 适用于全部代码、部分贡献另适用 Apache，不把 “Dual License” 自行理解为所有文件任选一份。
- `agent-client-protocol` 与 SDK 变体的正文证据继续有效。`markdown-it-py` 同时保留自身及上游 markdown-it 告知；`propcache`、`yarl` 有单独 NOTICE，后续汇编时不能只复制 LICENSE。
- Pillow `13.0.0` 的约 70 KiB 许可文件已有 22 个捆绑组件章节；不能只登记顶层 MIT-CMU，也不能因正文出现 GPL 字样便断言整个 Pillow 为 GPL。应将所带二进制与具体组件 / 可选许可逐项对应。
- LiteLLM `1.105.7` 的正文对字面 `enterprise/` 目录另作限制；当前 wheel 无该完整路径组件，只有 `enterprise_billing` 等名称，不能仅凭名称判定专有，也不能据此忽略其 Rust 原生部分。保留目录核对，后续绑定精确构建来源。

原始库存位于 `tmp/development-license-audit-20261031/`：

| 证据 | SHA-256 | 内容 |
| --- | --- | --- |
| `wheel-inventory.json` | `1a9efba33ac3f142bbc4f4e400a48405e855419b5038170977eb9f0e1bbe729c` | 119 项版本 / wheel 摘要、许可原文及路径、元数据、原生文件 |
| `wheel-summary.json` | `c14a11304f0de03efa054059ed7773b3951af6becfcd2479606020519ae5a5c2` | 不含长许可原文的索引 |
| `elf-dependencies.json` | `5f291554a1bbcc1373f181350f34c9f7cab9651b83e727f6d837378c63456fd7` | 63 个对象摘要及 `DT_NEEDED`；不覆盖静态链接闭包 |
| `docker-cli-buildinfo.txt` | `f42b510213b7d4c34767134330bcc537dc7735b50b93a10dc6247f8e2eba35da` | 未运行 CLI，通过 `go version -m` 读取的构建信息 |

### 缺口与处理方案

| 对象 | 现有缺口 | 建议处理与完成依据 |
| --- | --- | --- |
| `fastmcp-slim==3.4.11` | wheel 无许可文件，元数据声明 Apache-2.0；同版 `fastmcp` 已有正文 | 可以复用已有标准正文，但先核对 slim 的精确来源、版权和 NOTICE 需求；形成包名 / 版本 / wheel 哈希到告知的映射，不能仅复制兄弟包声明就算关闭 |
| `tokenizers==0.23.3` | wheel 无许可文件，仅 Apache 分类；另有原生 Rust 对象 | 补精确版本的上游 LICENSE / NOTICE、Rust 依赖与构建来源；同时解决顶层告知和二进制捆绑库，不只补一份通用 Apache 文本 |
| `regex==2026.9.29` | 许可文件指出 CPython 2.6 / 3.1 的 CNRI 来源，但只附 Apache 完整正文 | 补对应历史 CNRI 条款、版权及源码来源映射；不能用当前 PSF 标题替代其声明的历史许可 |
| `lupa==2.8` | 六个二进制含 Lua 5.1 至 5.5 与 LuaJIT 2.1；许可文件只列 Lupa / Lua | 核对每个捆绑解释器精确版本与版权，补 LuaJIT 对应告知；当前 [LuaJIT 官方 COPYRIGHT](https://raw.githubusercontent.com/LuaJIT/LuaJIT/v2.1/COPYRIGHT)可确认它有独立版权，但移动分支不能代替本 wheel 的来源证据 |
| `certifi==2026.10.23`、`tqdm==4.70.1` | 已有 MPL 告知 / MIT 文本及部分源文件，未核对完整覆盖范围和首选修改形式；certifi 的 Mozilla 来源链接指向移动 `tip` | 先确认 wheel 中哪些内容已经是源码形式；对不足部分绑定精确版本源码 / 生成材料及获取告知，保留原版权。建议随交付保存可验证源码，而非把“只给仓库首页”当作完成依据 |
| 26 个原生包，优先 cryptography / Pillow / cffi / hf-xet / tokenizers / LiteLLM / Lupa | 顶层许可不覆盖静态 Rust / C / C++ 闭包；ELF 动态依赖只能显示一部分 | 对每个精确 wheel 关联上游构建记录、锁与组件许可；Pillow 复用已带 22 章节，补版本映射。cryptography 的[官方说明](https://cryptography.io/en/latest/installation/#static-wheels)确认 wheel 可静态包含 OpenSSL，但不证明本轮精确 OpenSSL 版本；不得用宿主或基础镜像版本代填 |
| SDK 镜像的 Python / OS 基础层与 `pip==25.0.1` | 已固定 Python 镜像来源 digest，最终镜像可读，缺实际 OS 包版本 / 源包、版权、Python 捆绑库及 pip 内嵌依赖清单 | 从现有候选镜像静态提取包数据库、common-licenses、版权、Python LICENSE 和 pip vendor 信息；绑定实际包版本与源包获取方案。`pip` 的内嵌依赖不在 119 项运行锁中，必须单列 |
| Go 沙箱镜像 | 已有 `golang:1.26.7-alpine3.23` 的精确本地 ID，缺 Alpine 包与 Go / 标准库内嵌依赖告知 | 提取 APK 数据库、Alpine 许可、Go LICENSE / PATENTS / VERSION 和 vendor 清单，再按实际组件决定是否补来源 / 源码。不能给整镜像只标一个 Go BSD 许可 |
| Docker CLI `29.8.2` | 原归档无 LICENSE / NOTICE；已有二进制摘要。构建信息为 Go `1.26.8`、commit `7fc2dff`、CGO 静态外部链接，未提供依赖模块清单 | 若未来交付中携带 CLI，补对应源码 / vendor 告知及静态 C 运行库来源；目前只用于可信测试控制容器只读挂载，没有进入 SDK 镜像。按实际交付边界处理，不扩大到归档里未提取的 dockerd / containerd 等工具 |
| 统一交付告知 | 尚无覆盖最终镜像与捆绑依赖的完整第三方告知、源码映射及交付校验 | 证据齐后汇编包 / 组件 → 版本 / 哈希 → 版权 / 许可 / NOTICE → 源码位置的清单；按实际交付物保留正文并验证镜像内位置。现有 wheel 自带文件不能直接代替完整交付清单 |

这里区分内部技术试验与向使用者分发镜像。[Mozilla FAQ](https://www.mozilla.org/en-US/MPL/2.0/FAQ/)说明 MPL 对组织内部使用 / 分发与外部分发的要求不同；[MPL 正文第 3 节](https://www.mozilla.org/en-US/MPL/2.0/)约束覆盖代码的源码可得性和告知，不据此要求整个 RadishMind 开源。最终方式需按源码形式、可执行形式和实际接收者确认。Apache 的许可副本、修改说明、版权与适用 NOTICE 要求以[第 4 节](https://www.apache.org/licenses/LICENSE-2.0)为准；汇编告知不授予上游商标背书。

[Python 官方镜像说明](https://raw.githubusercontent.com/docker-library/docs/master/python/README.md#license)明确提醒镜像还有其他许可组件；[Python 3.12.16 的许可页](https://docs.python.org/3.12/license.html)也列有捆绑软件告知。两者只能提供审计原则，不能替代精确镜像清单。本机 `python:3.12-slim@aa943ce4…` 不能直接按该名称读取独立镜像，构建缓存中是否可独立导出未验证；未为补证拉取镜像。下一步从已存在的最终 SDK 镜像读取其基础层文件。

### 静态取证窗口方案（后续执行结果见下节）

先做无网络静态提取，再据实际清单拟定下载清单，避免一次授权无限范围的依赖下载：

1. 仅复用本地 SDK 镜像 `sha256:912f0a65e355629b9578c4c80d2da0b3b0881e467874b476dfdc3e3f816e4acf` 和 Go 镜像 `sha256:b17af760035fc2f338eed92d448a6c67f2d45438844fc6c60678fa5f99e44b57`，使用 `docker create --pull=never` 各创建一个临时容器，名称前缀 `radishmind-license-audit-20261031-`。配置无网络、只读根、无挂载 / 端口、capability 全移除、`no-new-privileges`，入口设为 `/bin/false`；不执行 `docker start / run / exec`。
2. 通过 `docker inspect / cp` 读取 SDK 容器的 `/etc/os-release` / `/usr/lib/os-release`、`/var/lib/dpkg/status`、`/usr/share/doc`、`/usr/share/common-licenses`、Python 的 `LICENSE.txt`、pip 的发行元数据与 `_vendor` 目录；读取 Go 容器的 `/etc/os-release`、`/lib/apk/db/installed`、`/usr/share/licenses`、`/usr/local/go/{LICENSE,PATENTS,VERSION}` 及 `src/vendor`。路径缺失记为缺口，不下载替代品、不猜版本。
3. 输出只写 `tmp/development-license-audit-20261031/image-evidence/`，总量上限 512 MiB，超过即停止并保留失败记录；每次 Docker 操作超时 60 秒。归档只作静态读取，不执行其中程序。按精确容器 ID 清理成功和失败路径，核对状态从未运行、资源清零；保留提取材料与摘要。
4. 不安装扫描器、不启动 Docker 守护进程、不连接管理 socket 到容器、不拉取 / 重建 / 推送镜像、不执行 SDK 或真实模型。此窗口只补证，不改变既有锁或宣布分发准入。

上述只读审计提交时，静态提取程序准备于 `tmp/development-license-audit-20261031/extract_image_licenses.py`，仅做语法检查，尚未执行。批准窗口后的命令为：

```sh
.venv/bin/python tmp/development-license-audit-20261031/extract_image_licenses.py
```

后续下载应分两类另报清单：一类为上述缺失正文 / 精确源码 / 原生构建清单；另一类为静态提取后实际发现的 OS 源包或捆绑组件。每项固定来源、版本、摘要或验证方式、体积上限及保存路径；目前尚缺 OS 清单，不能先声称来源集合已经完整。不因告知缺失直接降版、删除运行依赖或重建候选；若确需依赖调整，先说明影响并确认。

本次审计验证：119 个 wheel 摘要与冻结报告一致，所有已声明的 `License-File` 路径均存在；两项缺失包同时没有该声明，不能靠字段存在性检查发现。`./scripts/check-repo.sh` 全量通过（退出码 0，日志 `tmp/development-license-audit-20261031/check-repo-full.log`）；收尾文本、篇幅、中文规范、Markdown 链接及 `git diff --check` 复核通过。该次提交范围只包含审计结论与方案文档；当时静态取证程序仅在忽略目录准备并检查语法，没有执行，后续窗口尚未授权。

## 已授权的无网络镜像静态取证

起点为干净 `dev@1d2c175c`。项目所有者明确批准上述无网络窗口：复用现有 SDK / Go 镜像，创建但不启动临时容器，按既定路径提取并清理，再更新缺口与补下载清单，验证后提交 `dev`、不推送。本次没有新增下载、安装、构建、镜像拉取或任何容器内程序执行。

静态提取程序正常退出。两个容器均使用固定镜像 ID，实际配置无网络、只读根、无挂载 / 端口、非 root、capability 全移除和 `no-new-privileges`。取证前后均为 `created`，`Running=false`、PID 为 0、启动时间为零值；不是以进程退出码推定“没有执行”。容器已按精确 ID 删除，归属查询为空，没有补救清理。15 项路径中 14 项形成归档，合计 12317184 字节（约 11.75 MiB）；缺失项明确保留错误信息，未扩展路径或联网找替代品。

机器可读的[取证主记录](../../../scripts/eval/development_probe/sdk_variant/license-audit-20261031.json)固定镜像、结果、分片与原始证据摘要；原始归档、容器配置 / 状态、清理记录和解析清单位于 `tmp/development-license-audit-20261031/image-evidence/`。归档仅按成员读取，没有解包到系统路径或导入其中程序。所有归档摘要重新核对，Debian 文档链接只在已捕获的虚拟文件树内解析。

### 许可缺口的变化

| 范围 | 本次补齐的证据 | 仍需处理 |
| --- | --- | --- |
| SDK 基础 OS | `/usr/lib/os-release` 确认 Debian 13 / trixie、完整版本 13.7；87 个已安装包，归并为 61 个精确源包 / 版本；87 条版权路径全部可解析 | 保留 Debian 修订与 epoch，继续匹配实际源码 / Debian 补丁和获取方式；版权文件可能包含多文件许可，不按单个标题概括整个二进制包 |
| Debian 通用许可 | 已捕获 common-licenses 及各包版权，GPL / LGPL / MPL 等引用可继续逐项绑定 | 通用目录含某种许可不证明镜像有对应许可的组件；反之，版权文件存在也不等于对应源码义务和最终交付告知已经完成 |
| Python / pip | Python `3.12.16` 主 LICENSE、`pip==25.0.1` 自身 MIT 正文、完整 `_vendor/vendor.txt` 与内嵌文件已保留 | Python 捆绑库映射仍缺；pip 的 18 项内嵌依赖没有独立 LICENSE / NOTICE 文件，内联告知尚未逐项审完，需按内嵌精确版本补证，不能使用 119 项运行锁中的新版覆盖 |
| Go 基础 OS | APK 数据库确认 17 个包、10 个 origin；`alpine-release==3.23.5-r0`，每个 origin 有完整 aports commit 和版本 | `/usr/share/licenses` 缺失；`/etc/os-release` 只捕获到指向未取路径的符号链接，版本证据来自 APK 包而非该文件正文；按固定 aports 提取 APKBUILD、补丁与校验和，再补正文 / 源码 |
| Go 工具链 | `VERSION` 为 `go1.26.7`；主 LICENSE / PATENTS 和 `src/vendor` 四个模块的同类文件均在 | 已取 `x/crypto / x/net / x/sys / x/text` 的精确版本；本窗口没有取 `src/cmd/vendor`，不能声称全部 Go 编译工具内嵌依赖已覆盖；优先另行确认镜像内源码补取范围，确有不足再补对应发布源码 |
| 119 项运行依赖与测试 CLI | 原 wheel、二进制和许可审计证据保持，未改锁或镜像 | fastmcp-slim / tokenizers 告知、regex / Lupa 历史与捆绑正文、MPL 覆盖源码、26 个原生包构建映射、CLI 条件性交付及统一告知仍按前表处理 |

Alpine 声明可分为：`alpine-baselayout / apk-tools / busybox / pax-utils` 的相关包声明 GPL，`musl-utils` 是复合声明，CA 证书包声明 MPL / MIT，OpenSSL 为 Apache，另有 MIT / Zlib。此处是已捕获 APK 的声明，未以标签推断逐文件适用范围或直接得出侵权结论。两个镜像中的 OpenSSL 版本也不能用来代填 cryptography wheel 内静态 OpenSSL 的版本。

### 已更新的补下载清单（未授权、未执行）

所有精确版本和分组已进入主记录的三个分片，不把长清单复制回当前焦点：

| 清单 | 已固定的选择条件 | 下一步获取内容与验证 |
| --- | --- | --- |
| [Debian 源包清单](../../../scripts/eval/development_probe/sdk_variant/license-audit-20261031.parts/debian.json) | 61 个源包 / 精确版本，关联 87 个二进制包 | 先取 Debian 官方 archive / security / snapshot 的来源元数据和对应 `.dsc`；核对可信签名及 Sources / `.dsc` 摘要链，再列实际源码、Debian 修订 / 补丁的 URL / 大小 / 哈希，不以最新版代替。本表是保守取证候选集合，不宣称每个包都必须按同一方式提供源码 |
| [Alpine 来源清单](../../../scripts/eval/development_probe/sdk_variant/license-audit-20261031.parts/alpine.json) | 10 个 origin / 版本 / 完整 aports commit，关联 17 个二进制包 | 按固定 commit 取 APKBUILD、引用补丁和许可文件，保留来源及上游校验和。APK 数据库的 `Q1…` 是包校验信息，不冒充源码 SHA-256；不执行 APKBUILD，不安装 APK |
| [内嵌与 Python 来源清单](../../../scripts/eval/development_probe/sdk_variant/license-audit-20261031.parts/bundled.json) | pip 的 18 项内嵌版本、Go 的 4 个已见模块、Python `3.12.16`、Go `1.26.7`、29 个 Python 取证根项及条件性 Docker CLI `29.8.2` | 29 项为 26 个原生 wheel 加 fastmcp-slim / certifi / tqdm；先取对应发布元数据、官方来源版本 / 构建记录、LICENSE / NOTICE 和依赖锁。静态子依赖需从这些材料继续枚举，不声称 29 项就是最终闭包。Python 源码摘要已有镜像环境记录；其他源码摘要尚未取得，不能预填 |

优先顺序为发行物告知缺失与 pip 内嵌告知、Alpine 来源正文、Debian / Python / Go 精确来源映射，随后闭合原生捆绑依赖及最终告知。已有 18 项 pip 内嵌版本保留独立身份，例如 `certifi==2024.8.30`、`urllib3==1.26.20`、`setuptools==70.3.0`；它们与运行锁中的版本或包集合不同，不建议为许可取证直接升级 pip 或删除内嵌包。

建议下一步仅申请**来源元数据与告知下载**：总量上限 128 MiB、单文件 16 MiB，只写 `tmp/development-license-audit-20261031/downloads-stage1/`；限定上述官方来源和精确版本，逐项保存 URL、原始字节摘要与验证结果。只获取发布 / 构建元数据、固定版本来源文件、许可 / NOTICE、签名与校验材料；不下载大型源码归档、二进制 wheel / APK / DEB 或镜像，不执行 `setup.py`、APKBUILD、安装器或任何下载内容，不启动容器 / 服务。可信签名验证所需的信任依据应单列，不能把同时下载但未核验的公钥当作已可信。

该小批次完成后，再据真实元数据列出源码归档的精确地址、摘要、体积、补丁与生成材料，另行申请有界下载；定位失败、版本冲突或超限时停止对应项并记录缺口。此次静态取证授权不包含上述网络下载，也不授予重新构建、镜像发布或正式采用资格。静态窗口已完成并关闭；许可退出条件保持未完成。

本次收口验证：所有提取归档、原始清单、状态与记录分片的摘要一致；61 / 10 组来源分别与原始 87 / 17 条包记录逐项对应；两个容器的实际配置、零启动时间、PID 为 0 和清理结果复核通过。`./scripts/check-repo.sh` 全量通过（退出码 0，日志 `tmp/development-license-audit-20261031/check-repo-image-evidence-full.log`），Docker / SDK / 联合运行测试入口保持关闭；文档收尾后再次通过文本、篇幅、中文规范、链接及 `git diff --check`。提交仅含取证记录、清单分片与文档同步，不改依赖锁或执行配置，不推送。

## 已授权的来源元数据与告知下载

项目所有者批准上一节的第一阶段下载窗口：官方来源与精确版本，总量不超过 128 MiB、单文件不超过 16 MiB，不下载大型源码归档、不安装或执行下载内容。起点为干净 `dev`、HEAD `e4fa5da8`。本轮只写仓库隔离取证目录和审计文档，没有启动容器 / 服务、调用模型、改变依赖锁或构建镜像；网络窗口现已结束。

[来源取证主记录](../../../scripts/eval/development_probe/sdk_variant/license-sources-20261031.json)及其 `.parts/` 固定实际 URL、字节摘要、版本关联、源码候选与缺口；原始响应、逐请求账本、下载器和离线重建脚本保留在 `tmp/development-license-audit-20261031/downloads-stage1/`。总计 600 次请求，598 次成功，累计 15641241 字节（约 14.92 MiB），最大单文件 3958033 字节；失败响应也计入额度并保留摘要。下载器限制官方主机与 HTTPS 重定向、拒绝归档和二进制后缀，读取前预留额度；没有执行所取得的配方、源码文件或构建配置。

### 补证结果与尚存缺口

| 对象 | 本轮取得及核对 | 仍未完成 |
| --- | --- | --- |
| Debian 61 个源包版本 | snapshot 精确版本索引、文件信息与全部 `.dsc`；逐份核对 Source / Version、snapshot SHA-1 及文件体积，提取 SHA-256 与源码 / Debian 修订归档地址 | 本机没有 `gpg/gpgv`，也没有已建立信任的密钥链；没有安装工具或把下载公钥当作可信。签名未验证，候选不能据此视为可信源码交付闭包 |
| Alpine 10 个 origin | 全部固定 aports 提交配方的版本与 APK 数据库一致；130 个本地补丁 / 配置 / 来源文件按配方 SHA-512 核对通过 | 外部源码归档与两份 Debian netbase 文本未下载，体积待定；动态来源、完整构建输入和最终镜像告知仍待核对。没有执行 APKBUILD |
| Python 29 个取证根项及 pip 内嵌版本 | 共 48 份 PyPI 精确发布元数据；29 个实际 wheel SHA-256 均在同版发布记录中唯一匹配；每个版本列出一个 sdist 的 URL / SHA-256 / 大小 | sdist 未下载；上游发布元数据不能证明原生 wheel 的静态依赖、构建修订或生成材料闭包 |
| pip 25.0.1 | 官方源码树的 20 份内嵌告知按 Git blob 核对；18 项 vendor 版本与镜像清单一致，另取 typing_extensions 4.12.2 正文 | 镜像中的缺失告知尚未汇编补入；内联版权、pip 修改后的 vendor 源码仍待交付核对，原始上游 sdist 不能替代修改后来源 |
| fastmcp-slim / tokenizers | 取得精确源码树 LICENSE；slim 的 `pyproject.toml` 明确包名与 Apache-2.0，发布工作流已留存；tokenizers 的 Cargo 锁与构建配置已留存 | 还未形成包 → 实际 wheel → 完整告知 / 构建来源的交付映射；tokenizers 的 Rust 组件闭包未关闭 |
| Lupa 2.8 | 固定源码树、六个 Lua / LuaJIT 子模块提交与各自版权正文已留存，LuaJIT 2.1 来源不再只指向移动分支 | wheel 与各子模块的实际构建对应、Lua 5.1 来源仍待核对；源码树出现 LuaJIT 2.0 不等于该 wheel 携带它 |
| regex 2026.9.29 | 精确上游正文、CPython 2.6 / 3.1 历史正文已留存 | regex 声明 CNRI Python 1.6，而两份 CPython 正文列 CNRI 1.6.1；请求 `v1.6/LICENSE` 返回 404，已停止此项定位，不能静默替换或登记关闭 |
| Go 1.26.7 | 官方发布元数据给出源码归档摘要 / 体积；固定源码树告知、4 个 `src/vendor` 与另 12 个 `cmd/vendor` 模块已列明 | 上游清单不证明实际沙箱镜像已经包含全部告知；源码归档未下载，镜像对应关系与交付汇编仍待核对 |
| CPython 3.12.16 | 官方目录、源码发布 SPDX 清单、签名材料已取得；源码归档 SHA-256 与镜像环境值一致，SPDX 共 25 个包条目 | 签名尚未验证；源码发布清单不等于镜像或编译产物清单，实际捆绑 / 动态链接取舍仍待核对 |
| Docker CLI | 沿用本地只读构建信息，本轮未下载其来源 | 尚未确定是否随最终工具集交付，条件性缺口不扩张为默认下载范围 |

共核对 69 个 Git blob 的原始字节；所有成功及失败响应的已存字节重新计算 SHA-256。Alpine GitLab 首次请求返回 HTTP 418，保留错误响应后改用官方 `alpinelinux/aports` GitHub 镜像的同一完整提交；错误页没有进入配方结果。另一失败为上表 CPython 历史路径 404。未以其他版本或最新版替代缺失证据。

### 源码归档候选与下一窗口

| 候选组 | 已知文件与体积 | 前置条件 |
| --- | --- | --- |
| Debian | 132 个 `.dsc` 所列来源文件（含上游签名），313154899 字节 | 先建立可信签名核验路径；保留原始源包、Debian 修订与构建输入，不只下载 upstream orig |
| PyPI | 48 个 sdist，100108968 字节 | 区分 29 个 wheel 取证根项、pip 与其内嵌上游；从精确来源继续枚举原生构建材料，不把顶层归档当完整闭包 |
| Go | `go1.26.7.src.tar.gz`，34150794 字节 | SHA-256 来自官方发布元数据；优先判断已有源码证据能否满足实际交付需要 |
| CPython | `Python-3.12.16.tar.xz`，20839652 字节 | 官方 SPDX 摘要与镜像环境一致，但签名信任、构建及交付映射尚待完成 |
| Alpine | 固定配方已列外部地址及 SHA-512，体积尚未取得 | 不能加入可执行下载批次；先核定体积与完整输入，动态来源单列处理 |

前四组已知候选合计 468254313 字节（约 446.56 MiB），**不是当前授权额度，也不是必须全部下载的决定**；没有下载其中源码归档。清单是保守取证候选，不声明每项均须采用同一源码交付方式。下一步先收口信任核验、Alpine 体积及实际交付范围，再据精确清单申请新的有界下载；当前 128 MiB 窗口不顺延、不用于安装验证工具、重建或分发。最终第三方告知、覆盖源码提供方式与原生组件来源仍未闭合，阶段 0 许可退出条件保持未完成。

本轮验证：离线重建核对全部响应字节、分片摘要、29 个 wheel 发布摘要、61 份 `.dsc` 及 130 个配方输入，主记录与前轮取证记录的引用摘要一致。首次全量检查因系统 Go 缓存目录访问受限退出；改用仓库隔离 `GOCACHE`，以 `GOPROXY=off GOTOOLCHAIN=local ./scripts/check-repo.sh` 完整重跑通过（退出码 0，日志 `tmp/development-license-audit-20261031/check-repo-source-metadata-isolated-full.log`），没有为验证下载依赖。收尾文本、篇幅、中文规范、Markdown 链接与 `git diff --check` 再次复核通过。提交仅含来源证据索引 / 候选清单和文档，不推送。
