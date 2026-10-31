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

### 建议的下一取证窗口（尚未执行）

先做无网络静态提取，再据实际清单拟定下载清单，避免一次授权无限范围的依赖下载：

1. 仅复用本地 SDK 镜像 `sha256:912f0a65e355629b9578c4c80d2da0b3b0881e467874b476dfdc3e3f816e4acf` 和 Go 镜像 `sha256:b17af760035fc2f338eed92d448a6c67f2d45438844fc6c60678fa5f99e44b57`，使用 `docker create --pull=never` 各创建一个临时容器，名称前缀 `radishmind-license-audit-20261031-`。配置无网络、只读根、无挂载 / 端口、capability 全移除、`no-new-privileges`，入口设为 `/bin/false`；不执行 `docker start / run / exec`。
2. 通过 `docker inspect / cp` 读取 SDK 容器的 `/etc/os-release` / `/usr/lib/os-release`、`/var/lib/dpkg/status`、`/usr/share/doc`、`/usr/share/common-licenses`、Python 的 `LICENSE.txt`、pip 的发行元数据与 `_vendor` 目录；读取 Go 容器的 `/etc/os-release`、`/lib/apk/db/installed`、`/usr/share/licenses`、`/usr/local/go/{LICENSE,PATENTS,VERSION}` 及 `src/vendor`。路径缺失记为缺口，不下载替代品、不猜版本。
3. 输出只写 `tmp/development-license-audit-20261031/image-evidence/`，总量上限 512 MiB，超过即停止并保留失败记录；每次 Docker 操作超时 60 秒。归档只作静态读取，不执行其中程序。按精确容器 ID 清理成功和失败路径，核对状态从未运行、资源清零；保留提取材料与摘要。
4. 不安装扫描器、不启动 Docker 守护进程、不连接管理 socket 到容器、不拉取 / 重建 / 推送镜像、不执行 SDK 或真实模型。此窗口只补证，不改变既有锁或宣布分发准入。

对应静态提取程序已准备于 `tmp/development-license-audit-20261031/extract_image_licenses.py`，仅做语法检查，尚未执行。批准窗口后的命令为：

```sh
.venv/bin/python tmp/development-license-audit-20261031/extract_image_licenses.py
```

后续下载应分两类另报清单：一类为上述缺失正文 / 精确源码 / 原生构建清单；另一类为静态提取后实际发现的 OS 源包或捆绑组件。每项固定来源、版本、摘要或验证方式、体积上限及保存路径；目前尚缺 OS 清单，不能先声称来源集合已经完整。不因告知缺失直接降版、删除运行依赖或重建候选；若确需依赖调整，先说明影响并确认。

本次审计验证：119 个 wheel 摘要与冻结报告一致，所有已声明的 `License-File` 路径均存在；两项缺失包同时没有该声明，不能靠字段存在性检查发现。`./scripts/check-repo.sh` 全量通过（退出码 0，日志 `tmp/development-license-audit-20261031/check-repo-full.log`）；收尾文本、篇幅、中文规范、Markdown 链接及 `git diff --check` 复核通过。提交范围只包含审计结论与方案文档；静态取证程序仅在忽略目录准备并检查语法，没有执行，后续窗口尚未授权。
