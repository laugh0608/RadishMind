# 受控开发阶段 0：代表任务、SDK 与执行合同

更新时间：2026-11-12

状态：离线准备、人工任务复验、SDK LLM 子合同及单尝试持久事务 / 当前资格 / 内部通道固定响应验证已完成；受限 SDK Agent / 完整允许事件、工具往返及等待取消已通过固定响应验收；固定 Go 沙箱 / 受限工具 / 归属终止已真实验收；开发产物合同与 SQLite 显式保存 / 导出已实现并定向验证；真实 SDK 联合技术任务已通过；采用复评已完成但正式采用暂不通过：原生 HTTP 已通过固定上游技术验证，阶段 1 合同范围与 A 级画板已确认，许可交付仍待闭合。

所属功能：[受控软件开发团队](../features/workflow/controlled-software-development-team-v1.md)。本卡集中记录阶段 0 探针、单尝试控制合同与工具边界，不为各层再建任务卡或检查器。

## 2026-11-12 收尾与下次接续

所有者已确认[仅四项工具元数据的新方案](../../scripts/eval/development_probe/sdk_variant/pillow-tool-metadata-plan-20261112.json)，并要求今晚先提交、回顾近期提交和更新文档后结束。本次不执行获取；下次复核原输入、核对器 / 解析器指纹和九项离线测试后，可在原四请求 / 16 MiB 正文 / 80 秒网络 / 10 分钟总窗 / 32 MiB 留存范围内执行，无需重复确认同一范围。方案文件保留编制时的 `prepared_not_authorized` 快照，本段及[收尾记录](../devlogs/2026-W46.parts/2026-11-12.md#2026-11-12-近期提交复核与收尾)登记后续确认，不能把确认当作已获取。

十九份源码继续复用；四项元数据仍未请求。成功后先复核构建器 / 工具实物、依赖闭包与资源预算，不自动追取。任何失败按原约定结束全窗；镜像层、容器、工具安装、构建、发布仍另行确认，许可交付与正式采用复评尚未完成。

## 2026-11-10 规划收口接续

本卡继续记录原有技术结果。单 Agent [公开合同](../contracts/development-task-delivery-v1.md)及六幅代表面的修订要求已确认。原画板修订出图已获项目所有者最终视觉确认。许可仍需按具名来源缺口裁决，并完成一个最终候选的交付验收。具体退出条件和阶段 1 后的真实单 Agent 价值复核见[实施专题](../features/workflow/controlled-software-development-team/implementation-and-evaluation-plan-v1.md)。

11 月 11 日完成 Pillow 单 wheel 候选首轮六次输入取证：五份 blob 精确匹配，multibuild 快照按 Git 类型 / 可执行位、路径、长度和内容摘要完整核对，见[结果](../../scripts/eval/development_probe/sdk_variant/pillow-fixed-inputs-result-20261111.json)与[周志](../devlogs/2026-W46.md#2026-11-11-pillow-首轮取证与交付阻塞复核)。本窗口已结束；构建镜像 / 工具、嵌套源码和最终交付未完成，保留单一候选，阶段 0 不退出。

同日获准执行四次输入解析，取得 AOM / dav1d 完整提交和 cibuildwheel 精确源码包内的构建器镜像摘要 / 工具约束，见[结果](../../scripts/eval/development_probe/sdk_variant/pillow-input-resolution-result-20261111.json)。窗口已结束；尚未拉取或验证构建镜像实物，源码归档和工具闭包仍缺，未构建或修改运行锁。

随后获准的源码 / 镜像描述符 / 工具元数据下载窗口，在第二次请求返回 config HTTP 302 时按零重定向规则立即停止，见[停窗结果](../../scripts/eval/development_probe/sdk_variant/pillow-material-acquisition-result-20261111.json)。manifest 摘要已匹配，列出 23 层 / 572782550 字节压缩体积；config、平台与层实物未验，19 份源码与四项工具元数据均未请求。具名 CDN 接续范围须另行确认，阶段 0 仍不退出。

同日获准一次具名 CDN 跳转后，config 摘要、`linux/arm64` 声明与 23 层 rootfs 对应通过；取得 XZ / zlib-ng 两份源码并完成静态路径、版本头与指纹核对。第七次请求在 xcb-proto TLS 连接超时后按约定停窗，未重试或换源，见[接续结果](../../scripts/eval/development_probe/sdk_variant/pillow-material-continuation-result-20261111.json)。剩余 17 份归档与四项工具元数据待恢复窗口；镜像层、工具实物及最终交付仍未验收。

随后获准的 xcb-proto 一次有界重试已执行：四份已验材料本地复算匹配；原 URL 首请求在 10.100353 秒返回 curl `56` 与 TLS 协议版本告警，HTTP `000`、正文 0 字节，见[重试结果](../../scripts/eval/development_probe/sdk_variant/pillow-xcb-retry-result-20261111.json)。已全窗停止，没有后续请求、换源或 TLS 参数调整；剩余清单不变，不将本次告警归因为上次连接超时，也不推断归档不存在。网络处置需结合现有证据另定，镜像层和容器运行仍另行确认。

显式 SOCKS5 代理尝试及所有者手动访问也未取得归档后，按要求核查官方来源。已直接取得 2024-04-15 发布公告中的 gz / xz 摘要；已验四份材料复算匹配。[替代计划](../../scripts/eval/development_probe/sdk_variant/pillow-xcb-alternative-plan-20261111.json)选择目录可访问且列出目标的 GWDG Gentoo 镜像，以原版 `1.17.0.tar.xz` 和公告双摘要替代旧 gz 获取项，尚未下载或授权执行。该路径不是 X.Org 原始发布路径，后续配方须显式适配 xz；其它归档、镜像层和阶段退出结论保持。

随后所有者确认执行替代计划，xcb-proto 实际取得 151748 字节，SHA-256 / SHA-512 与公告一致；64 项成员、`1.17.0` 版本声明、许可及构建要求已核对。原 libxcb URL 随后返回指向 HTTP 的 301，按 HTTPS 限制未跟随并停窗，见[替代执行结果](../../scripts/eval/development_probe/sdk_variant/pillow-xcb-alternative-result-20261111.json)。本轮两次请求、151843 字节；已验源码累计三份，剩余 16 份及四项工具元数据。没有自行改写跳转、换源或下载镜像层。

后续只读核查已取得 libxcb `1.17.0` 官方公告双摘要，并与 Gentoo / FreeBSD 的 xz 大小及摘要记录交叉匹配；GWDG 目录与精确文件 HEAD 响应头列出 455208 字节。[新替代计划](../../scripts/eval/development_probe/sdk_variant/pillow-libxcb-alternative-plan-20261111.json)固定同版本 `.tar.xz` 镜像 URL，五份已验材料复算后直接复用，首项验过再接续 15 份归档及四项元数据。当前仅方案准备，未取得 libxcb 正文、未变更配方或运行锁；镜像层及容器运行仍另行确认。

同日获准执行后，libxcb 的 455208 字节与公告双摘要一致，静态核验通过；随后原清单中的 libjpeg-turbo / zstd / TIFF 也完成归档结构、版本及许可 / 构建声明核对。libavif 在一次允许的 HTTPS 跳转后返回 HTTP 200，但只收到 7836809 / 14665927 字节即达到该项总时限、curl 退出 28，立即全窗停止，见[执行结果](../../scripts/eval/development_probe/sdk_variant/pillow-libxcb-alternative-result-20261111.json)。九次请求合计 17348190 字节；已验源码累计七份，剩余 12 份完整归档和四项元数据。部分 libavif 正文保留但不计验收，未重试、续传或继续后项；镜像层、构建和阶段出口状态不变。

11 月 12 日按要求完成[离线复核](../../scripts/eval/development_probe/sdk_variant/pillow-material-recheck-20261112.json)：manifest / config 与七份归档摘要匹配，所选文本与归档原文一致；历史文本累计比留存文本多 96719 字节，原因尚未完整确定，新窗口仍按较大历史占用扣预算。[libavif 新方案](../../scripts/eval/development_probe/sdk_variant/pillow-libavif-retry-plan-20261112.json)保持原 URL / 版本与代理，拟将本项含跳转时限改为 240 秒，完整核验后接续 11 份归档及四项元数据。方案待授权；本轮零网络请求，无镜像层、容器或构建操作，阶段 0 状态不变。

随后所有者确认执行新获取方案。九份复用材料及计划 / 回执绑定复算匹配；libavif 完整 14665927 字节与旧部分正文前缀一致，libpng 完整取得并静态验过。lcms2 归档标称 2.19.1，但 configure.ac / CMake / Meson 均声明 2.19，按约定全窗停止，未计源码验收；七次请求合计 21978198 字节，见[执行结果](../../scripts/eval/development_probe/sdk_variant/pillow-libavif-retry-result-20261112.json)。已验源码累计九份，尚有含 lcms2 在内的十份未验收及四项工具元数据；没有继续后项、换源、重试或镜像层 / 容器操作，阶段 0 不退出。

随后按要求复用 lcms2 归档并查阅官方来源：[版本复核](../../scripts/eval/development_probe/sdk_variant/pillow-lcms2-version-review-20261112.json)确认 ChangeLog 含 2.19.1 修补节，且三个版本声明文件及 ChangeLog 的 Git blob 全部匹配发布页所指提交 `21c582a594fe5279f90c0b93437c398f93bf62b0`。建议分别登记发布版本 2.19.1、源码声明 2.19，接受此精确摘要归档的具名差异；仅提出裁决建议，尚未确认采纳，验过源码仍计九份。未证明整个归档与 Git 树一致、未验证发布签名或恢复后续下载。

所有者随后确认接受该精确 lcms2 归档的具名差异，[复核记录](../../scripts/eval/development_probe/sdk_variant/pillow-lcms2-version-review-20261112.json)已追加接受决定，分别保留发布版本 2.19.1、源码声明 2.19、头文件值 2190 和归档摘要；不改写旧停窗回执。十二份材料复算匹配，已验源码累计十份。[剩余获取方案](../../scripts/eval/development_probe/sdk_variant/pillow-remaining-acquisition-plan-20261112.json)限定九份归档和四项工具元数据，31 次请求 / 272 MiB 正文 / 620 秒网络时间 / 45 分钟总窗口，任一失败或超限停窗。版本复核额外文本已纳入静态预算；本轮零网络请求，下载、镜像层和容器运行均待另行确认。

随后按所有者授权执行剩余获取：十二份复用材料与引用复算匹配；OpenJPEG 2.5.4、libwebp 1.6.0、brotli 1.2.0 完整取得并核对结构、版本、许可和构建声明。freetype 原 URL 返回 302 至允许域名外的 ftp.cc.uoc.gr，未跟随并立即停窗，见[执行结果](../../scripts/eval/development_probe/sdk_variant/pillow-remaining-acquisition-result-20261112.json)。六次请求共 7099413 字节，已验源码累计十三份；剩余六份归档和四项元数据。未换源、重试、跳过继续或进行镜像层 / 容器操作，阶段 0 不退出。

随后按要求复核 freetype 来源：十五份复用材料与本地 302 回执摘要匹配；FreeType 官网指向 Savannah，Savannah 文档确认该入口分配镜像，GNU 官方清单列有 ftp.cc.uoc.gr 的 GNU 镜像路径。结合本地精确 Location，已准备[新方案](../../scripts/eval/development_probe/sdk_variant/pillow-freetype-mirror-plan-20261112.json)：直接请求该精确 HTTPS 地址一次、零跳转，核验后接续五份归档与四项元数据。网页工具未能读取目标目录，未取得目标正文或独立摘要；当前仅计划准备，下载 / 镜像层 / 容器待另行确认。

随后所有者确认执行精确镜像方案。十五份复用材料与引用复算匹配；freetype 唯一请求在 8.005493 秒返回 curl `28`、`SSL connection timeout`，HTTP `000`、正文 0 字节，见[执行回执](../../scripts/eval/development_probe/sdk_variant/pillow-freetype-mirror-result-20261112.json)。立即停止全窗，后续五份归档及四项工具元数据未请求；没有重试、换源、调整协议或跳过。验过源码仍为十三份，剩余六份归档与四项元数据；镜像层及容器未操作，阶段 0 状态不变。来源候选依据保持，本次超时不证明目标归档不存在或具体网络根因。

随后按要求仅复核本地超时材料：十五份复用文件、回执与执行脚本绑定匹配；CONNECT 200 后 curl 自身 8 秒连接时限失败，执行器未主动终止，未达到总时限。缺少分阶段时间与代理上游证据，无法在链路、代理转发、镜像响应或 TLS 交互之间确定根因。[有界获取方案](../../scripts/eval/development_probe/sdk_variant/pillow-freetype-bounded-retry-plan-20261112.json)保持同 URL / 代理和零跳转，首项拟调整为连接 20 秒、总计 90 秒，增加有限 verbose 与阶段计时；验过后才接续原清单，全窗仍 20 次请求 / 208 MiB 正文，网络上限改为 470 秒。方案待授权；本轮未探测镜像或恢复下载，镜像层及容器仍另行确认。

随后所有者确认执行该有界方案。十五份复用材料与引用匹配；同一 freetype URL / 代理的唯一请求在 20.005775 秒返回 curl `28`、`SSL connection timeout`，源站 HTTP `000`、正文 0 字节，见[执行回执](../../scripts/eval/development_probe/sdk_variant/pillow-freetype-bounded-retry-result-20261112.json)。诊断记录连到本地代理 TCP 为 0.000379 秒、CONNECT 200、发出 ClientHello；未记录 ServerHello / TLS 完成，不能据此区分代理上游、链路或镜像根因，也不将 `ssl_verify_result=1` 单独判作证书错误。立即全窗停止，未请求后续五份归档和四项元数据；未重试、换源、调整协议或继续加时。十三份已验源码及阶段 0 状态保持，镜像层和容器未操作。

随后按要求只读核查官方备用来源与成本。FreeType 官网列出 SourceForge 稳定发布渠道，2.14.3 目录 README 给出原 tar.gz 的 SHA-256，并列有 833 字节签名文件；未获取签名 / 公钥或验签。[新获取方案](../../scripts/eval/development_probe/sdk_variant/pillow-freetype-sourceforge-plan-20261112.json)固定原格式与发布摘要，拟从 SourceForge FRS 入口获取，最多两次具名主机 / 精确路径内跳转；首项 8 MiB / 20 秒连接 / 90 秒全过程，验过后接续原清单，全窗 22 次 HTTP / 184 MiB / 470 秒网络时间。十五份既有材料及 Pillow sdist 复算匹配；[路线成本复核](../../scripts/eval/development_probe/sdk_variant/pillow-route-cost-review-20261112.json)建议仅调整获取渠道、有条件继续单 wheel 路线，构建暂缓。23 层压缩总量 572782550 字节可确认，展开体积 / 工具闭包 / 编译成本仍未知。方案及建议待确认，本轮没有归档、签名、镜像下载或容器运行，阶段 0 不退出。

所有者随后确认执行 SourceForge 方案。freetype 取得 4134916 字节并匹配发布 SHA-256；harfbuzz、bzip2、AOM、dav1d、libyuv 五份也完成结构、版本 / 提交 URL 绑定及许可 / 构建声明核对，见[执行回执](../../scripts/eval/development_probe/sdk_variant/pillow-freetype-sourceforge-result-20261112.json)。八次 HTTP 共 33962305 字节，源码累计十九份。工具元数据核对脚本在联网前导入 packaging 失败，立即停窗，四项元数据均未请求；未安装依赖、修复后重跑或继续后项。已验成果保留，构建器实物、工具闭包、最终交付与阶段 0 状态不由源码齐备推定通过。

随后按要求修正元数据核对器：改为独立的本地 JSON 输入 / 标准输出入口，显式使用仓库 pip 内嵌 packaging 并留存源码指纹，无安装或导入回退；目标标签固定为 CPython 3.12 / manylinux_2_28_aarch64。九项离线正反例通过，十九份源码及 manifest / config 共二十一份输入复算匹配，见[预检记录](../../scripts/eval/development_probe/sdk_variant/pillow-tool-metadata-preflight-20261112.json)。[新计划](../../scripts/eval/development_probe/sdk_variant/pillow-tool-metadata-plan-20261112.json)只含原四项 PyPI JSON，四次请求 / 16 MiB 正文 / 80 秒网络 / 10 分钟总窗 / 32 MiB 新留存，任一失败停窗。当前零联网、方案待确认；镜像层、容器、构建和阶段 0 出口保持原边界。

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
