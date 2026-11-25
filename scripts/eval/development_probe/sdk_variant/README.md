# SDK 遥测可选依赖候选

`1.59.0+radishmind.1` 于 2026-10-31 完成变体构建、119 项冻结安装和固定响应联合复验。原始 `1.59.0` 的 138 项锁和镜像保留为历史试验基线。2026-11-21 最终 ARM64 SDK 采用修复后的 Pillow wheel，通过禁网构建、告知读取和 SDK / 联合回归。2026-11-25 确认本机材料可取得后，按阶段 1 固定 Go 单 Agent 的受限执行范围正式采用。[最终交付记录](#最终-sdk-本地交付)

原变体准备器与构建校验覆盖六包 14 份 wheel 补充告知、pip 25.0.1 的 21 份内嵌告知和 CPython 3.12.16 的 Expat / libmpdec 两份告知。wheel 校验绑定安装报告中的版本与归档摘要。pip 独立校验同一安装的元数据及 280 个内嵌文件。CPython 校验绑定保留镜像的 `pyexpat` 和 `_decimal` 扩展字节。下文 10 月 31 日的历史镜像早于补充告知的接入。最终配方使用[独立离线上下文](#最终-sdk-交付离线准备)。

## 来源与改动

`source.json` 固定 PyPI `1.59.0` 源码归档 URL / SHA-256、精确上游文件摘要、补丁摘要和修改后文件摘要。许可证来自上游 `v1.59.0/LICENSE`，Git blob 也有记录。`optional-laminar.patch` 只有以下改动：

- 改为独立本地版本 `1.59.0+radishmind.1`；声明 MIT 并把原许可证与修改记录放入发行物。
- 将 `lmnr` 从必需依赖移至 `laminar` 可选安装项；不改版本范围，不修改或分发专有代理代码。
- 显式启用 Laminar / OTel 观测但缺少依赖时，初始化和直接创建根 span 均明确报错；默认关闭时沿用上游延迟导入和直接执行路径。

`prepare_sdk_variant.py` 只消费本地归档，不联网、不安装、不构建、不导入 SDK。它只创建仓库 `tmp/` 内的新目录，拒绝错误摘要、路径逃逸、重复路径、链接与特殊文件；先检查补丁再应用，核对结果摘要，移除副本中的旧发行元数据并记录完整构建输入摘要。准备器还通过 `sdk_notice_delivery.py` 校验并复制清单中的补充告知：精确 wheel 必须匹配已验冻结锁，正文与目标路径必须一致；缺失、篡改、路径越界、重复或未登记正文均拒绝。失败保留不完整目录，不覆盖已有证据。

CPython 告知放在镜像的 `/usr/local/share/radishmind/third-party/cpython/`。Expat 保留 `Modules/expat/COPYING` 原文。libmpdec 从 `Modules/_decimal/libmpdec/mpdecimal.c` 复制完整首段许可注释，仅追加一个末尾 LF。源码证据见[工具链清单](license-toolchain-image-result-20261105.parts/python.json)。Dockerfile 现有校验入口会同时检查两份正文和对应扩展的精确字节。

`build-requirements.lock` 仅锁定构建工具；原 138 项锁派生的 `constraints.txt` 仅约束既有版本，不强制安装未被使用的包。正常 pip 解析决定新的依赖闭包；只有明确构建的本地 SDK wheel 可使用本地 URL，其余发行物必须为有 SHA-256 的 PyPI wheel，出现 `lmnr` 或代理包即拒绝生成候选锁。实际生成的独立锁为 [requirements-linux-arm64-py312.lock](requirements-linux-arm64-py312.lock)，运行解析、全部 wheel 与安装报告的版本 / 摘要已逐项核对。

`--no-deps` 仅用于单独构建 SDK wheel。依赖解析、下载和最终安装均保留依赖检查，最终执行 `pip check`；没有安装后删包。最终镜像的冻结安装阶段禁用网络，只读取导出的 wheel 集合与锁。构建配方会将补充告知与其部分范围清单复制到 `/usr/local/share/radishmind/third-party/`，再使用 pip 安装报告检查版本、wheel 摘要及告知字节；不一致时构建失败。目录 / 文件按 `0755` / `0644` 准备，供镜像内非 root 用户读取。告知的实际安装与可读性以 2026-11-21 的最终镜像验收为准。

## 已验候选与证据

精确来源、wheel / 锁 / 镜像摘要、移除包清单、测试与清理记录见 [acceptance-20261031.json](acceptance-20261031.json)，实验过程见[周志](../../../../docs/devlogs/2026-W44.parts/development-sdk-variant-20261031.md)。

- 环境为 `linux/arm64` / Python `3.12.16`；119 项解析与安装集合不含 `lmnr` / 专有代理，较原锁减少 19 项；除本地 SDK 变体外，其余包的版本和 wheel 哈希全部保持。镜像另含基础镜像自带 `pip`，所以元数据枚举为 120 项。
- wheel 中的 MIT 正文、修改记录和被修改的 SDK 源文件与 `source.json` 摘要一致；最终安装禁用网络，`pip check` 通过。镜像实际 ID 为 `sha256:912f0a65e355629b9578c4c80d2da0b3b0881e467874b476dfdc3e3f816e4acf`。
- 4 项遥测可选依赖检查、7 项 LLM 子合同及完整 SDK 工作进程的工具往返、失败、未知用量、等待取消与父存活管道退出通过。
- 联合正常 / 旧测试失效 / 命令中取消三场景通过；保存失败后重试、幂等、重启重读、严格导出及独立权限撤销均覆盖。SDK 身份降权与管理通道 / 控制数据访问拒绝通过；所属容器清零，没有宿主补救清理。

本次锁固定的是已构建 wheel 的确切字节；重新从源码打包不保证 ZIP 时间戳相同。复验可复用保留的 wheel 集合并核对本锁；重新解析 / 打包必须比较版本与摘要，不能悄悄覆盖已验锁。记录不宣称镜像逐位可复现。

## 依赖与运行窗口命令

以下为本次已授权并执行的范围，窗口现已结束；后续执行需另行授权并使用新隔离目录 / 运行编号，不能覆盖现有证据。只在 `linux/arm64`、既有 Docker 环境中执行，不启动 Docker 守护进程，不安装宿主机或工作台依赖。隔离根目录为 `tmp/development-sdk-variant-20261031/`；构建阶段允许固定源码、PyPI wheel 和固定基础镜像访问，运行阶段无外网、无宿主端口。构建缓存、镜像、wheel、报告和失败材料保留，所有本轮运行容器已删除；不推送或发布。

### 固定源码、解析与冻结安装

从仓库根目录执行。源地址和哈希由 `source.json` 给出；先下载到新文件，再由准备器核验。首次构建导出 wheel 与解析报告；第二次构建只消费该冻结集合。构建使用 Docker 现有资源配额，容器运行上限另见下节。

```sh
variant_root="$PWD/tmp/development-sdk-variant-20261031"
curl --fail --location --max-time 120 \
  -o "$variant_root/openhands_sdk-1.59.0.tar.gz" \
  https://files.pythonhosted.org/packages/77/2c/3c6a59e4088deeae25bf7255844e53374c337e69ec2446c0549df269af92/openhands_sdk-1.59.0.tar.gz
python3 scripts/eval/development_probe/prepare_sdk_variant.py \
  --source-archive "$variant_root/openhands_sdk-1.59.0.tar.gz" \
  --directory "$variant_root/context"
docker build --platform linux/arm64 --target artifacts \
  --output "type=local,dest=$variant_root/context/resolution" "$variant_root/context"
docker build --platform linux/arm64 --network none --target runtime \
  --iidfile "$variant_root/image-id.txt" \
  -t radishmind-development-sdk-variant:1.59.0-radishmind.1 "$variant_root/context"
```

构建成功后，检查 `resolution.json / requirements.lock / source.json`、wheel 许可证与镜像实际安装集合；实际本地 SDK wheel 摘要必须与解析、锁和安装报告相同。使用镜像真实 digest 做后续运行，不依赖可移动 tag。若缺失依赖、构建失败或摘要不一致，保留失败并停止，不放宽约束重试。

### 镜像内正反例与 LLM 子合同

以下每条命令保存独立日志。每轮检查实际容器配置、退出状态和 OOM；即使命令失败也删除对应精确容器，不能只依赖成功路径的 `--rm`。

```sh
set -e
variant_image="$(cat "$variant_root/image-id.txt")"
for check in optional_laminar contract inspect; do
  name="radishmind-sdk-variant-$check"
  script="/opt/probe/test_sdk_$check.py"
  if [ "$check" = inspect ]; then script=/opt/probe/inspect_sdk.py; fi
  docker create --name "$name" \
    --network none --read-only --cap-drop ALL --security-opt no-new-privileges \
    --tmpfs /tmp:size=64m,mode=1777 --memory 1g --memory-swap 1g --cpus 2 --pids-limit 128 \
    "$variant_image" python "$script" > "$variant_root/$check.container-id.txt"
  trap 'docker rm -f "$name"' EXIT
  docker inspect "$name" > "$variant_root/$check.container.json"
  status=0
  docker start --attach "$name" > "$variant_root/$check.log" 2>&1 || status=$?
  docker inspect --format '{{json .State}}' "$name" > "$variant_root/$check.state.json"
  docker rm "$name"
  trap - EXIT
  test "$status" -eq 0
  python3 -c 'import json,sys; s=json.load(open(sys.argv[1])); assert s["ExitCode"] == 0 and not s["OOMKilled"]' "$variant_root/$check.state.json"
done
```

可选依赖测试覆盖精确版本 / MIT 告知、默认环境无 `lmnr` / 代理分发包、可选依赖 marker、SDK 导入、关闭遥测的正常调用、四个显式遥测环境入口与直接创建根 span 的拒绝。LLM 测试继续覆盖真实 SDK 工具往返、非法输入、错误不重试、超时及缺失 usage；固定响应只在容器回环临时端口监听。

### 完整 SDK 与联合任务复验

本轮用现有 Go 缓存准备并执行了 `sdk-check.test`、`joint/joint.test` 和 `joint/code-runner`；Linux Docker CLI 复用此前保留文件，不重新安装。SDK 层沿用[原工作进程验收命令](../README.md#受限-agent-工作进程验收2026-10-31)，替换本轮镜像 digest、二进制挂载路径和独立容器名称，维持非 root、无外网、只读根、1 GiB / 2 CPU / 128 进程、256 MiB 临时目录，覆盖完整事件、失败、等待取消和父管道退出。

联合任务使用已生成的 `joint/input/list-filter-plan.json`，其中候选版本必须为 `1.59.0+radishmind.1`，其余三文件摘要 / 工具 / 预算保持；不改写历史冻结计划。复用已验固定代码镜像与 Docker CLI：

```sh
socket_path="$(docker context inspect --format '{{.Endpoints.docker.Host}}')"
socket_path="${socket_path#unix://}"
python3 scripts/eval/development_probe/run_joint.py \
  --directory "$variant_root/joint" --socket "$socket_path" --run-id 1 \
  --sdk-image "$variant_image"
```

此步需要明确授权可信控制容器访问本机 Docker 管理 socket。权限、Linux 私有挂载父目录与 SDK 降权拒绝检查沿用[联合拓扑](../README.md#sdk代码沙箱与产物联合验收2026-10-31)：控制容器仅增补必要五项 capability，无外网、无端口、只读根，1 GiB / 2 CPU / 128 进程和 512 MiB 临时空间；代码容器继续独立无网络 / 非 root。覆盖候选修改 / 测试 / 保存 / 导出、旧测试失效和命令中取消。结束必须核对归属容器为空，管理 socket、原工作区和数据库不得进入 SDK 或代码沙箱。

## 采用限制

只有新闭包、冻结安装、无专有代理的发行物检查及上述实际 SDK / 联合回归均通过，才将本变体登记为技术候选。启用 `laminar` 可选安装项会重新引入原许可问题，不属于本镜像或当前授权范围。MPL 对应源码、其他依赖及基础镜像的分发义务仍需闭合；本轮不接真实提供方、不实现正式 Web、不宣布 SDK 采用或阶段 0 退出。

许可补证进度：已按单独授权静态提取现有 SDK / Go 镜像文件，两个容器始终未启动且已清理。实际 OS 包、内嵌依赖、缺失路径和后续下载候选见 [license-audit-20261031.json](license-audit-20261031.json) 及其分片；镜像清单本身不代表来源已取齐或已获准分发。后续已按明确授权完成有界来源元数据 / 告知下载，结果与源码候选见 [license-sources-20261031.json](license-sources-20261031.json)；后续已按独立授权完成[来源归档与签名链核验](license-archives-20261031.json)：61 个 Debian 源包通过仓库签名链绑定，9 项 Alpine 来源体积 / SHA-512 已核对，191 项候选字节完整取得并核对。下载与隔离验签窗口已关闭，10 个验证容器已清理。2026-11-05 已从保留材料生成[补充告知](notice-supplements/README.md)与[精确映射](license-delivery-20261105.json)，覆盖 fastmcp-slim、tokenizers、Lupa 与 cryptography 内嵌 OpenSSL 的 11 份正文；仅准备仓库材料，未安装到镜像。原生组件构建映射、完整告知交付与最终覆盖源码交付仍未闭合；不自动下载新发现的依赖或重建镜像。

本轮随后按单独批准的八个地址取得[来源元数据](license-metadata-20261105.json)：六项 HTTP 200，两项历史 Python 1.6 地址为 404，总正文 1357427 字节。cryptography 精确 wheel 内的 OpenSSL SBOM 已绑定 4.0.3 源码 URL / SHA-256 / 构建参数，PyPI 发布声明的 wheel 摘要一致；尚未取得该源码归档、验证签名信任或确认历史构建容器摘要。Rust SBOM 标记 `all_targets=true`，不能作为仅已链接组件清单。窗口已关闭，不自动追加下载。

[四项来源补证窗口](license-source-followup-20261105.json)已获准并完成，结果见[源码核对记录](license-source-result-20261105.json)：五次 HTTP 请求共 55282876 字节，OpenSSL 4.0.3 归档摘要与 wheel SBOM 一致、许可正文与仓库副本相同。Python 1.6 归档候选为 404，官方发行页已取得；返回的两个 Git 标签均为 alpha，不能作为最终版。下载窗口已关闭，未追取发行页链接。随后另获批准执行[两项官方页面补证](license-page-followup-20261105.json)：下载说明页取得最终 1.6 八条全文，FAQ 为 404，总正文 69066 字节。已补充 regex 所声明的 CNRI Python 1.6 正文，现为五包 12 份、48485 字节；与既有 1.6.1 文本的版本、年份、handle 和第 7 条差异明确记录，见[页面结果](license-page-result-20261105.json)。正文缺失已补齐，regex 继承来源 / 完整告知范围仍待核对；未点击接受、提交表单或下载源码，窗口已关闭。

补充告知交付代码已通过离线正反例测试，并与摘要已复核的历史候选安装报告匹配。此结果只证明构建材料与已安装 wheel 的对应关系；实际新镜像安装、完整原生组件清单与覆盖源码交付仍待验收。

### 原生组件离线对应

[原生组件记录](license-native-20261105.json)将 26 个已锁定 wheel、63 个 AArch64 ELF、现有许可正文及 10 份内嵌 SBOM 对应，原文件摘要与冻结锁 / 既有 ELF 证据一致。Pillow 的 18 个捆绑共享库可由 23 条动态链接、SONAME 与 ORIGIN 搜索路径作静态对应；其余动态链接集中在八个外部 SONAME。文件名、声明版本和系统包存在分别记录，不冒充运行时实际加载。

七份 Rust SBOM 带 `all_targets=true`，其中 required / excluded 条目不能直接等同最终静态链接清单。Pillow 的 optional 项不自动计为已交付；其 OpenJPEG ancestor 的名称 / purl 冲突、zlib 声明版本与外部 `libz.so.1` 的边界、组合许可正文中的 AOM / DAV1D 等组件均保留为待核对项。原生对象与 SBOM 没有被导入或执行。

[静态镜像取证](license-native-image-result-20261105.json)已单独获准并完成：25 次复制共 8353280 字节，八个库的 SONAME、原路径 / 链接目标、正文及 dpkg 清单 / MD5 声明均匹配；对应 glibc、GCC、zlib 三组已留存源码版本。库自身未引出新动态依赖名称。容器始终处于未启动状态，已按精确 ID 删除并复核归属容器为空；未启动 Docker、拉取 / 重建镜像或执行原生程序，窗口关闭。该结果不替代实际加载、静态组件、独立构建来源与最终告知 / 覆盖源码交付验证。


### 基础镜像 pip 告知

`prepare_sdk_variant.py` 同时准备 wheel 告知与 `third-party/pip/`。后者由 `pip_notice_delivery.py` 消费精确 pip 证据，校验 21 份正文和 280 个内嵌文件指纹。候选 Dockerfile 在无网络步骤读取实际 pip 元数据和完整 `_vendor` 集合，版本、重复元数据、缺失 / 新增 / 篡改文件或符号链接均拒绝。源码内容只读校验，不导入 pip；离线测试不代替镜像重建验收。取证与当前状态见[清单](license-pip-notices-20261105.json)和[说明](pip-notice-supplements/README.md)。

### 工具链与历史来源补证

[工具链结果](license-toolchain-image-result-20261105.json)记录已授权的八次镜像复制、80 个 ELF 静态对象及 Go `cmd/vendor` 853 个文件的源码对应，两个容器始终未启动且已清理。[历史来源结果](license-historical-provenance-result-20261105.json)记录 27 次 HTTP 200、11 份源码文本和发布声明摘要核对；发布证明不含构建材料，本轮未验证签名信任。所有窗口已关闭，原计划的准备状态不作为当前执行结论。新增外部库、内嵌告知、历史输入及最终交付仍有缺口；完整判断见[框架采用评审](../../../../docs/features/workflow/controlled-software-development-team/framework-and-gateway-integration-v1.md#2026-11-05-框架采用评审)。

### Pillow 工具元数据离线核对

`../pillow_tool_metadata.py` 只读取本地 PyPI JSON，向标准输出写候选记录，错误返回非零；没有网络、安装或构建入口。显式使用仓库 `.venv` 中已有的 `pip._vendor.packaging`，不自动回退或安装依赖。该内嵌 API 仅用于当前批次，先保存解析器版本 / 源码指纹并完成预检；环境变化时重新核对，不视为新的长期依赖约定。

从仓库根目录执行离线预检：

```sh
.venv/bin/python -B scripts/eval/development_probe/pillow_tool_metadata.py --preflight
.venv/bin/python -B -m unittest discover -s scripts/eval/development_probe -p test_pillow_tool_metadata.py -v
```

取得明确授权并留存真实响应后，才以 `--tool setuptools|pybind11|meson|ninja --input <本地响应路径>` 核对。目标为 CPython 3.12.16 / manylinux_2_28_aarch64，保留原始依赖 marker；声明为空则如实记录，不解析完整依赖闭包。标签匹配不等于构建器实物、工具安装或 Ninja 二进制版本已验。[预检记录](pillow-tool-metadata-preflight-20261112.json)与[四项获取计划](pillow-tool-metadata-plan-20261112.json)分别记录离线证据和编制时的窗口快照。所有者已于 11 月 12 日确认该方案，按收尾安排留到下次执行，见[任务卡](../../../../docs/task-cards/development-task-probe-v1-plan.md#2026-11-12-收尾与下次接续)；旧失败回执不改写。

### Pillow 离线 before-all 绑定

暂存项目的 `.radishmind/pillow/` 包含 `pillow-before-all.sh`、`native-build.sh`、`pillow-test.sh` 和对应的哈希清单。将已验输入目录只读挂载到 `/opt/radishmind/pillow-inputs`。`CIBW_BEFORE_ALL_LINUX` 从 `/project` 调用 `bash .radishmind/pillow/pillow-before-all.sh`。将脚本的 `--cibw-environment` 输出设为 `CIBW_ENVIRONMENT_LINUX`，使后续的构建和修复复用 `before-all` 的搜索路径。脚本要求固定的 CPython 3.12.15，在新虚拟环境按哈希安装 Meson 和 Ninja，再调用本地 `native-build.sh`。

原生配方须把已核验的原生输入安装到 `/opt/radishmind/pillow-native`，库目录为 `lib`。`pkg-config` 使用 `/usr/bin/pkg-config`，先查前缀下的 `lib/pkgconfig`、`share/pkgconfig`，再查 `system-pc` 中显式放行的镜像 `xau.pc` 及其 `Requires: xproto` 对应的 `xproto.pc`。检查要求 `xcb-proto` 和 `xcb` 的 `.pc` 版本为 1.17.0，`libxcb.so`、zlib-ng 兼容模式的 `libz.so` 均解析到前缀内且 SONAME 对应 `libxcb.so.1`、`libz.so.1`。LibXau 使用构建器 `/usr/lib64/libXau.so.6.0.0`。实际 wheel 的 ELF 依赖在修复后核对。

测试阶段将 `--cibw-test-environment` 输出设为 `CIBW_TEST_ENVIRONMENT_LINUX`，从已验 RPM 展开的 `/opt/radishmind/pillow-inputs/fribidi/usr/lib64` 加载 FriBidi。将 `CIBW_TEST_COMMAND_LINUX` 设为 `cd {project} && bash .radishmind/pillow/pillow-test.sh`，由脚本把已验图片放入 `Tests/images/` 并运行上游测试命令。`pillow-test-requirements.lock` 记录 Pillow `tests` extra 与 NumPy 的完整哈希闭包。`native-build.sh --prepare-only ARCHIVES NEW_DIRECTORY` 只校验哈希、解包、应用补丁并提取告知。本地暂存目录、文件指纹和主控限制均已记录。[实物预检回执](pillow-offline-build-inputs-result-20261119.json) · [测试来源锁定记录](pillow-offline-test-sources-20261119.json)

### Pillow 候选的伴随告知与 SDK 复验

`pillow-notice-supplements/` 内的五份补充告知与 `13.0.0+radishmind.1` wheel 摘要绑定。使用 `sdk_notice_delivery.verify` 对照目标 SDK 中真实离线安装的 pip 报告和告知正文。已有冻结锁仍用于识别目标镜像的原始安装，复验将候选安装到容器的临时目录。正式交付时须将伴随告知纳入最终镜像的交付清单并复核实际可读性。[告知清单](pillow-notice-supplements/manifest.json) · [许可来源](pillow-notice-sources-20261121.json) · [SDK 复验](pillow-sdk-recheck-result-20261121.json)

### 最终 SDK 交付离线准备

从仓库根目录准备一个新的 `tmp/` 上下文。准备器核对历史锁、119 个原始 wheel 和替换 wheel 的摘要，只把最终 119 个 wheel、配方及告知写入新上下文。生成的 `delivery-inputs.json` 记录冻结锁、配方、源码索引和逐个 wheel 的摘要。

```sh
python3 scripts/eval/development_probe/prepare_sdk_delivery.py \
  --wheelhouse tmp/development-sdk-variant-20261031/context/resolution/wheels \
  --pillow-wheel "tmp/pillow-build-window-20261119.7zt4q4x4/run-03/output/pillow-13.0.0+radishmind.1-cp312-cp312-manylinux_2_27_aarch64.manylinux_2_28_aarch64.whl" \
  --directory tmp/sdk-delivery-context-new
```

获准构建后，以本地已验的固定基础镜像及 `linux/arm64` 运行新 Dockerfile。构建器的 registry 访问须在构建窗口内关闭，步骤使用 `--network=none`。wheelhouse 作为只读构建挂载使用。pip 安装报告与完整冻结锁及告知由 `65532:65532` 校验。完成后按实际镜像摘要在禁网、只读根容器内再次核对告知和源码获取说明，并执行目标 SDK 与联合验收。伴随源码按实际交付镜像摘要及接收方获取路径复验。[源码清单](sdk-delivery-sources-20261121.json) · [预检回执](sdk-delivery-preflight-20261121.json) · [验收条件](../../../../docs/task-cards/development-task-probe-v1-plan.md#2026-11-21-最终-sdk-交付离线预检)

### 最终 SDK 本地交付

`deploy/Dockerfile.development-sdk-delivery` 使用独立离线上下文和固定基础镜像。最终镜像已核对 119 项冻结安装、`pip check`、`65532:65532` 读取 42 份补充告知、CPython 动态库归属及三个 SDK 联合用例。[构建与运行回执](sdk-delivery-build-runtime-20261121.json) · [联合验收回执](sdk-delivery-runtime-closeout-20261121.json)

已按最终镜像清单摘要将本地配套源码与告知放在 `dist/sdk-delivery/sources/<SDK-image-sha256>/`。目录内的 `manifest.json` 绑定镜像、冻结锁与全部材料，可用随附的 `verify_manifest.py` 离线逐项核对。交付记录保留 OpenSSL 补丁清单未证实的边界，历史构建取证已结束。[组装与回读回执](sdk-delivery-local-source-result-20261121.json) · [接收确认与采用复评](../../../../docs/task-cards/development-task-probe-v1-plan.md#2026-11-25-sdk-接收确认与阶段-0-采用复评)
