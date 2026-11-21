# 受控开发阶段 0：代表任务、SDK 与执行合同

更新时间：2026-11-19

状态：离线准备、人工任务复验、SDK LLM 子合同及单尝试持久事务 / 当前资格 / 内部通道固定响应验证已完成；受限 SDK Agent / 完整允许事件、工具往返及等待取消已通过固定响应验收；固定 Go 沙箱 / 受限工具 / 归属终止已真实验收；开发产物合同与 SQLite 显式保存 / 导出已实现并定向验证；真实 SDK 联合技术任务已通过；采用复评已完成但正式采用暂不通过：原生 HTTP 已通过固定上游技术验证，阶段 1 合同范围与 A 级画板已确认，许可交付仍待闭合。

所属功能：[受控软件开发团队](../features/workflow/controlled-software-development-team-v1.md)。本卡集中记录阶段 0 探针、单尝试控制合同与工具边界，不为各层再建任务卡或检查器。

## 2026-11-18 候选选型

19 份原生源码、构建器 manifest / config、计划和引用摘要均匹配，解析器版本及 20 份源码指纹一致，9 项离线测试通过。依序请求四项 PyPI JSON 后，setuptools 83.0.0 的 wheel 摘要与构建锁匹配，pybind11 3.1.0 和 Meson 1.12.1 各有一个符合目标声明的 wheel。Ninja 返回 HTTP 200，但在单项 20 秒到期时仅留存 290816 / 380593 字节，监控器终止 curl，当即停窗。四次请求共留存正文 629844 字节，含部分 Ninja 响应，没有重试或续传。[逐项回执、候选版本与摘要](../../scripts/eval/development_probe/sdk_variant/pillow-tool-metadata-result-20261118.json)

首窗结束时因缺少完整 Ninja 响应而暂停。随后获准单独向同一 PyPI 地址重新 GET 一次。新窗内 HTTP 200、curl 正常退出，正文与 Content-Length 同为 380593 字节。离线核验确认 Ninja Python 分发包 1.13.2 有一个匹配目标标签的 aarch64 wheel，声明大小 177180 字节。内含二进制版本待实物核验。[Ninja 单项回执及 wheel 摘要](../../scripts/eval/development_probe/sdk_variant/pillow-ninja-metadata-result-20261118.json)

**四个构建工具 wheel 已逐项核验。** 每个精确地址各 GET 一次，四次均为 HTTP 200。四包共 2613340 字节，大小和 SHA-256 均与预选记录匹配，786 条 `RECORD` 记录核对通过。下表为包内顶层许可正文及与 PyPI 声明一致的依赖条数。[逐包许可文件、依赖声明和摘要](../../scripts/eval/development_probe/sdk_variant/pillow-tool-wheel-result-20261118.json)

| 工具 | 实收字节 | 顶层许可正文 | `Requires-Dist` 条数 |
| --- | ---: | --- | ---: |
| setuptools 83.0.0 | 1008090 | MIT | 50 |
| pybind11 3.1.0 | 319402 | BSD-3-Clause | 1 |
| Meson 1.12.1 | 1108668 | Apache-2.0 | 4 |
| Ninja 1.13.2 | 177180 | Apache-2.0 | 0 |

55 条依赖声明均带可选功能（extra）条件。Ninja 的 `METADATA` 没有 `Requires-Dist` 字段，PyPI 对应值为 `null`。setuptools 另含 12 个内嵌分发包，其中 autocommand 附 LGPLv3 正文，配置目录的两份 NOTICE 涉及 MPL-2.0 和 BSD-3-Clause。Meson 的顶层许可与 COPYING 一致，`mesonbuild/dependencies/python.py` 首行 SPDX 标识写为 `Apache-2.`，已按包内原文留痕。

### 固定构建器的 23 层取证方案

已用留存的 manifest 和 config 复算固定 `manylinux_2_28_aarch64` 镜像摘要，逐项对齐 23 个压缩层描述符、23 个 `rootfs.diff_ids` 和 42 条 history 中的 23 条文件系统记录。第 4、15、16 个位置引用相同的 32 字节对象，压缩摘要和 DiffID 均相同。取证覆盖全部 23 个位置，按摘要只取 21 个对象；最大对象为第 22 层，274790712 字节。[逐层摘要、取证目标和预算明细](../../scripts/eval/development_probe/sdk_variant/pillow-builder-layer-plan-20261118.json)

| 预算项 | 核定值及计量方式 |
| --- | --- |
| 下载 | 23 个位置按序相加为 572782550 字节。冷缓存逐个获取唯一对象的正文为 **572782486 字节（546.25 MiB）**。另给鉴权、跳转和错误响应正文合计 1 MiB。 |
| 新增磁盘 | 建议静态取证峰值上限 **1 GiB**。压缩对象占 572782486 字节，路径索引含日志和临时文件最多 192 MiB，定向证据最多 128 MiB，请求回执与运行日志最多 8 MiB，余下 157026410 字节留给文件系统分配开销。串行写入的部分对象计入对应对象额度，同卷改名。 |
| 解压与空间 | 在镜像外流式读取 gzip/tar，最多两遍，每遍累计解压流量上限 8 GiB。只落盘索引及定向证据。执行前目标卷至少可用 2 GiB，运行中保留 1 GiB；达到限额即停止。实际展开大小和未来构建峰值均待实测。 |

| 待核对象 | 静态验收依据 |
| --- | --- |
| 解释器与 ABI | 解析 `/opt/python/cp312-cp312` 的链接终点，把构建目标和 PATH 中的 `python3`、`pip` 分开。以 ELF 架构、`patchlevel.h`、`pyconfig.h` 与 sysconfig 核对 AArch64、3.12 补丁版本及扩展 ABI。 |
| 系统与打包工具 | 核对 GCC/G++ 与 Clang 的实际搜索顺序、binutils、Make、CMake（libavif 至少 3.22）、Autotools、pkg-config、Meson、Ninja，以及补丁和解档工具。另对照 cibuildwheel 的 CPython 约束核查 build、auditwheel、patchelf 等入口与包归属，四个已验 wheel 单独标为外部输入。 |
| 系统库和原生依赖 | 核对 glibc 与动态加载器、C/C++ 运行库及启动对象、LibXau/XCB、系统 zlib、FriBidi 加载项、Tcl/Tk 及 CPython 实际引用的库。对照 19 份原生源码区分系统提供与候选自建的同名头文件、`.pc` 和库，尤其是 zlib-ng 的兼容构建与 SDK 侧的 `libz.so.1`。 |

### 第 1 层传输中断

执行前目标卷可用 300975013888 字节，满足 2 GiB 起始要求。首个对象经 quay.io 的一次 302 跳转至获准的 cdn01.quay.io 摘要路径。CDN 返回 HTTP 200，Content-Length 为 71996907 字节，实际只收到 18874368 字节，curl 以退出码 18 结束。两次 HTTP 共计 1104.77 秒，未触及单对象 45 分钟上限。按方案立即停窗，新增磁盘占用观测峰值 19062784 字节，低于 1 GiB。[请求回执、部分文件指纹与未验位置](../../scripts/eval/development_probe/sdk_variant/pillow-builder-layer-result-20261118.json)

| 核验项 | 本窗实物证据与缺口 |
| --- | --- |
| 层对象 | 第 1 层的 18 MiB 部分正文留存在独立目录，缺 53122539 字节，完整对象 SHA-256 与 DiffID 均待核。其余 20 个唯一对象未请求。通过核验的唯一对象为 0/21，层位置为 0/23。 |
| 解释器与 ABI | 最终视图无法形成，CPython 和引导 Python 的链接终点、版本头、sysconfig 与 ELF 架构均待核。 |
| 工具 | 编译器、链接器、原生构建工具、打包后端和 wheel 修复工具的镜像内路径、版本与归属均待核。此前验过的四个外部 wheel 只记录为构建输入。 |
| 库与头文件 | glibc、启动对象、系统库及与 19 份源码同名的库缺少最终可见成员；头文件、SONAME、动态依赖、覆盖关系和 RPM 归属均待核。 |

当前暂停层获取，等待可核对摘要的完整对象来源，或另行确认的传输方案。本窗没有重试、续传、运行容器或构建。CPython 在 SDK 中的动态库归属、其它具名来源和最终许可交付仍待核定。

### 2026-11-19 接续决定

离线复算了方案、manifest、config、停窗回执和 18 MiB 残片的摘要。在已检查的本地留存中未找到完整层对象。上次 CDN 正文平均约 16.70 KiB/s；仅按 90 分钟网络上限计算，传完 21 个唯一对象至少需要平均 103.58 KiB/s，尚须留出连接与跳转时间。保留单一 Pillow 候选，优先等待本地完整压缩对象；若改用网络，先确认有实质改善的来源与传输预算。

可接收固定摘要的 OCI 目录 `blobs/sha256/`，或按 SHA-256 十六进制值命名的原始压缩对象目录。对象就地只读，逐个核对大小和压缩摘要；齐备后再验证 23 个层位置的 DiffID，按原方案形成最终文件视图。重新压缩所得对象不能替代原始压缩摘要。已有 manifest、config、19 份源码和四个工具 wheel 可复用。本轮仅完成离线接续预检，实物验收仍为 0/21 个对象、0/23 个层位置。

#### 直连元数据核对与拟获取窗口

第 1 层经原代理及直连、第 22 层经直连均在 `quay.io` 返回 HEAD 200。三次响应的 `Docker-Content-Digest` 和 `Content-Length` 与固定描述符一致，直连第 1、22 层分别耗时约 0.84、0.92 秒。HEAD 正文共 0 字节，未收到 `Accept-Ranges`，GET 路径与实际正文吞吐仍待验证。[OCI Distribution Spec](https://specs.opencontainers.org/distribution-spec/) · [RFC 9110 §14.2](https://www.rfc-editor.org/rfc/rfc9110.html#section-14.2)

拟从 `quay.io` 直连按原摘要串行 GET。第 1 层从零开始，计入连接及可能的一次跳转，在 16 秒内收齐首 2 MiB 才继续同一响应，检查点相当于 128 KiB/s。失败即停窗并只保留本次残片和回执。首层及后续各唯一对象的大小、压缩 SHA-256 和 DiffID 逐个核验，全部通过后才形成最终文件视图。

拟延用单对象 45 分钟、累计网络 90 分钟、目标卷至少 2 GiB 可用、新增占用上限 1 GiB；最多 42 次 HTTPS，请求对象正文共 572782486 字节，其他响应正文上限 1 MiB。每对象最多一次具名 `cdn01.quay.io` HTTPS 跳转并核对摘要路径。无鉴权、重试、续传或代理回退。旧残片留存原处，本次只使用独立临时目录，不安装工具、启动容器或构建。待项目所有者授权实际正文获取。

#### 直连执行与最终文件视图

所有者随后批准本次直连窗口。首层在 4.295 秒收齐 2 MiB，沿同一响应继续；21 次原站 GET 均返回 302，21 次具名 CDN GET 均返回 200。实收层正文 **572782486 字节**，累计网络 191.913 秒，新增占用观测峰值 652251136 字节（约 622.0 MiB）。21 个唯一对象的大小、压缩 SHA-256 与全部 23 个层位置的 DiffID 均匹配。按层序合并后可见 77554 条路径，定向读取 265 条路径，RPM 数据库解析 338 个包记录。原残片独立留存，本窗已结束。[执行回执](../../scripts/eval/development_probe/sdk_variant/pillow-builder-layer-direct-result-20261119.json) · [定向文件证据](../../scripts/eval/development_probe/sdk_variant/pillow-builder-layer-direct-result-20261119.parts/static-findings.json)

| 核验项 | 实物与剩余差异 |
| --- | --- |
| 系统与解释器 | AlmaLinux 8.10。`/opt/python/cp312-cp312/bin/python` 指向 `/opt/_internal/cpython-3.12.15/bin/python3.12`，ELF 为 AArch64，版本头为 3.12.15，SOABI 为 `cpython-312-aarch64-linux-gnu`。系统 `/usr/bin/python3.12` 是另一个 RPM 文件。按既有目标 3.12.16 复评补丁版本差异。 |
| 编译与配置 | PATH 优先使用 GCC/G++ 14.2.1、binutils 2.41；Make 4.2.1 由 RPM 归属核对。CMake 入口指向 pipx，4.4.4 有 `dist-info` 路径及原生 ELF 摘要，运行版本待核。已定位 `/usr/local/bin` 的 Autoconf、Automake 和 Libtool，版本依据目前为镜像历史及文件路径。`/opt/clang` 未定位到工具文件。 |
| Python 工具 | 目标解释器内的 pip 26.2.1、build 1.6.1、packaging 26.3、pyproject-hooks 1.3.3，以及独立 pipx 环境中的 auditwheel 6.8.2、patchelf 分发包 0.19.1.0 和 pyelftools 0.33 已核对。PATH 的 pkgconf 1.4.2 与约束 `3.0.7.post0` 有差异，未定位到约束中的 virtualenv 21.14.6。Meson、Ninja 与 pybind11 的已验 wheel 为外部构建输入，配方仍须明确各工具的实际入口及离线供给。 |
| 系统库与原生输入 | 核对了 glibc 2.28、启动对象与主要 C/C++ 库。镜像已有 LibXau 1.0.9、XCB 1.13.1、zlib 1.2.11 和 Tcl/Tk 8.6.8；已绑定可见路径及对应 RPM。配方还须绑定自建 XCB 1.17.0、zlib-ng 等输入的实际头文件、`.pc` 与链接目标。目标 CPython 扩展声明 `libssl.so.1.1`、`libffi.so.6` 和 Tcl/Tk 8.6 的 ELF 依赖，已定位同名文件及归属 RPM。定向路径未定位 FriBidi/raqm，`libgcc_s.so.1` 和 `libcrypt.so.1` 的实际文件尚缺 RPM 归属。 |

固定对象获取与静态视图已验收。选型继续保留这一固定候选，先裁决 CPython 补丁差异，核定实际工具与原生构建输入，再决定是否进入构建窗口。容器、工具执行、候选构建和最终许可交付仍待各自验收。

#### CPython 兼容性与离线配方复评

**保留固定构建器作为单 wheel 候选，当前暂缓进入构建窗口。** 复算了 Pillow 13.0.0 源码包、发布工作流和 cibuildwheel 4.3.1 精确源码。按 `cp312-manylinux_aarch64` 的默认 `build` 前端、`pinned` 约束和启用的 Pillow 测试复核实际调用。[复评回执与精确 wheel 元数据](../../scripts/eval/development_probe/sdk_variant/pillow-builder-compatibility-review-20261119.json)

| 核验项 | 裁决与依据 |
| --- | --- |
| CPython 3.12.15 → 3.12.16 | 两端为 AArch64，SOABI 均为 `cpython-312-aarch64-linux-gnu`。固定构建器的已装头文件和 3.12.16 精确源码中有 213 个同路径头文件，212 个逐字节相同，唯一差异为 `patchlevel.h`。历史 Pillow wheel 的 8 个扩展共引用 108 个 CPython 动态符号，包括 `_PyBytes_Resize` 等私有符号，全部可在留存的 3.12.16 SDK `libpython` 找到。静态结果支持沿用 3.12.15 编译。新产物修复后须验证 `cp312-cp312-manylinux_2_28_aarch64` 标签、ELF 依赖，以及在固定 3.12.16 SDK 中的安装、导入和 Pillow 功能。 |
| pkgconf | Pillow 的 `setup.py` 读取 `PKG_CONFIG`，默认调用 `pkg-config`。镜像 `/usr/bin/pkg-config` 解析到 RPM 提供的 `/usr/bin/pkgconf` 1.4.2，满足 libxcb 生成脚本的 0.9.0 版本比较，`xcb-proto >= 1.17.0` 是独立的 `.pc` 模块要求。cibuildwheel 4.3.1 的 Android 分支显式安装 `pkgconf==3.0.7.post0`，本 Linux 路径直接复用镜像工具。配方应绑定该入口，并确定自建 XCB、zlib-ng 与允许复用的系统库对应的 `.pc` 搜索范围。 |
| virtualenv | Linux 默认 `build` 前端先执行隔离构建，测试阶段才由目标解释器的 pip 通过 `-c /constraints.txt` 安装 `virtualenv==21.14.6`，随后使用内嵌 pip 和 `--no-periodic-update` 创建测试环境。精确版本闭包含 virtualenv 本身和 `distlib 0.4.3`、`filelock 4.1.0`、`packaging 26.3`、`platformdirs 4.12.4`、`python-discovery 1.6.2`。本地仅有可按原摘要复用的 packaging wheel，另外五个 wheel 的发布元数据已核对，声明合计 6162608 字节，实物待取得。 |
| 原生构建工具 | Linux `before-all` 把 `/opt/python/cp39-cp39/bin` 放在 PATH 首位，镜像实物是 3.9.25。原依赖脚本通过该 `python3` 安装 Meson 和 Ninja，已验 Meson 1.12.1 要求 Python ≥3.10。离线配方须把这两个工具的安装及调用显式绑定到兼容的固定入口，本路线采用构建器的 CPython 3.12.15 作为该配方设计的入口。 |

测试 pip 应保留 `constraints-python312.txt` 的版本约束，通过 `PIP_NO_INDEX=1` 和 `PIP_FIND_LINKS` 使用已验摘要的本地 wheelhouse。现有测试约定还包含 Pillow 的 `tests` extra、NumPy、FriBidi 和测试图片。原脚本仍有 `yum` 和可变归档下载，原生配方也仍需接入固定归档和工具 wheel。进入构建窗口前，先取得缺少的 virtualenv wheel，固定测试输入和离线供给，并让配方明确解释器、`.pc` 搜索与链接目标。实际构建和 SDK 运行验收留待有边界的运行窗口。[CPython ABI 稳定性](https://docs.python.org/3.12/c-api/stable.html) · [3.12.16 变更记录](https://docs.python.org/3.12/whatsnew/changelog.html#python-3-12-16-final)

#### 离线工具与测试来源锁定

按复评记录逐个取得五个 wheel，共 **6162608 字节**，与留存的 packaging 26.3 一起核对。六个 wheel 的大小、SHA-256、231 条 `RECORD` 记录、许可文件和 CPython 3.12.15 Linux 上生效的依赖均通过。distlib 为 PSF-2.0，virtualenv、filelock、platformdirs 和 python-discovery 为 MIT。virtualenv 的第三方告知和四个内嵌 pip / setuptools wheel 也单独登记。[逐包回执](../../scripts/eval/development_probe/sdk_variant/pillow-virtualenv-wheel-result-20261119.json)

| 项目 | 已固定的输入和边界 |
| --- | --- |
| FriBidi | AlmaLinux 8.10 AppStream 的 `fribidi-1.0.4-9.el8.aarch64.rpm`，89632 字节，SHA-256 已由官方 `repomd.xml` 指向的 primary 元数据核对。仅用于 Pillow 测试加载 `libfribidi.so.0`，RPM 实物、签名和库路径待取得后核验。 |
| 测试图片 | `python-pillow/test-images` 提交 `073ab5ff1895abf8cf5c8d797c1066ed394ce207`，Git 树 `c62a7f86bd84166a002332db858b020361dfeea0`。已重算 Git 树，锁定 78 个 blob 的路径、模式、字节数和 Git 摘要，共 68635563 字节。按提交获取归档后逐项核验，限定为离线测试输入。 |
| `before-all` 与链接 | 新入口校验 `/opt/python/cp312-cp312/bin/python` 为 3.12.15，用它建立独立工具环境并按哈希安装 Meson 1.12.1 和 Ninja 1.13.2。原生前缀为 `/opt/radishmind/pillow-native`，`.pc` 先查前缀 `lib/pkgconfig`、`share/pkgconfig`，再查 `system-pc` 中镜像的 `xau.pc` 和它声明的 `xproto.pc`。配置锁定 XCB 1.17.0 和 zlib-ng 兼容库的链接目标，LibXau 复用镜像实物。独立测试环境从测试输入目录加载 FriBidi。 |

**当前仍不具备进入构建窗口的条件。** `before-all` 已有可审查入口，仍需准备 `/project/.radishmind/pillow/native-build.sh`，接入已验原生归档及补丁，并替换原配方的联网入口。FriBidi RPM、测试图片归档及 Pillow `tests` extra 与 NumPy 的离线材料仍待取得或核验。完成这些输入的准备后再核定单 wheel 构建、修复和测试窗口。本次检查覆盖 Bash 语法、环境输出、工具哈希及镜像路径。[测试来源锁定记录](../../scripts/eval/development_probe/sdk_variant/pillow-offline-test-sources-20261119.json) · [离线入口说明](../../scripts/eval/development_probe/sdk_variant/README.md#pillow-离线-before-all-绑定)

#### 原生配方与离线实物预检

| 项目 | 复核结果 |
| --- | --- |
| FriBidi | 89632 字节的 RPM 匹配锁定 SHA-256。校验了 OpenPGP RSA/SHA-256 主头签名，并核对签名主头中的压缩 payload 摘要。签名子钥指纹为 `E53CF5EF91CEB0AD1812ECB851D6647EC21AD6EA`。测试目录已准备 AArch64 `libfribidi.so.0`、链接终点和原包许可文件。 |
| 测试图片 | 57430295 字节的提交归档已留存，78 个 blob 逐项核对相对路径、字节数、可执行位与 Git 摘要。测试目录只用于 Pillow 验收。 |
| Python 测试 | Pillow `tests` extra 和 NumPy 2.5.5 共 17 个精确 wheel。按 CPython 3.12.15 / Linux AArch64 解析，最终目录用 `--no-index --require-hashes --dry-run` 全部解析成功。合并构建工具及 virtualenv 后，离线 wheelhouse 共 25 个不同工件。 |
| 原生配方 | `native-build.sh` 直接消费 19 份锁定归档，按 Pillow 原配方的 `-p0` 应用 sdist 内的 OpenJPEG 补丁。AOM、dav1d、libyuv 和 libwebp 显式使用本地源码。保留 FreeType → HarfBuzz → FreeType 次序，并从对应归档收集原生告知。`--prepare-only` 已验归档、解包、补丁和告知生成。`pillow-test.sh` 使用本地 FriBidi 和图片，保留上游三个测试命令。 |

**容器内构建输入与静态配方已备齐，实际构建窗口仍未开启。** 已组装 `13.0.0+radishmind.1` 的本地暂存项目和供只读挂载的输入目录。运行前须确认 `cibuildwheel 4.3.1` 主控和依赖的离线供给，确认选定引擎已具备精确 `linux/arm64` 镜像，并核定禁网资源与清理范围。仓库 `.venv` 当前无可导入的 cibuildwheel。真实编译、修复和 wheel 测试留待该窗口验收。[实物与预检回执](../../scripts/eval/development_probe/sdk_variant/pillow-offline-build-inputs-result-20261119.json) · [AlmaLinux 签名密钥](https://almalinux.org/security/#gpg-keys) · [签名子钥指纹](https://wiki.almalinux.org/release-notes/8.8)

#### 禁网构建窗口预检

| 项目 | 实测结果 |
| --- | --- |
| 宿主主控 | 获准串行获取 7 个 wheel，共 438341 字节。与 4 个已锁 wheel 合并后，11 个工件的大小、SHA-256、许可文件、312 条 `RECORD` 和 Darwin arm64 / CPython 3.14.5 依赖闭包均已核验。pip 隔离、禁网、强制哈希的试解析成功。宿主工具及其依赖只供本地构建使用。 |
| 固定镜像 | OrbStack 起始无该镜像。获准从 21 个本地已验压缩层流式导入，耗时 9.088 秒。导入后的 `linux/arm64` 镜像 ID 与 Quay manifest 原摘要一致，23 个有序 DiffID 与原 config 逐项匹配，可按完整摘要查询。保留本批唯一标签供构建使用。 |
| 构建命令 | 已验 sdist 中 71 个包内文件与 cibuildwheel 4.3.1 wheel 对应文件一致。直接从这些源码及依赖 wheel 运行 `--print-build-identifiers`，只得到 `cp312-manylinux_aarch64`。实际主控对本机摘要解析出 `--platform=linux/arm64 --pull=never`。配置解析确认显式 `--pull=never`、`--network=none`、关闭 `/host` 挂载，并仅只读挂载已验输入。 |

**预检通过，可申请一次禁网构建窗口。** 本次授权完成了依赖获取和镜像导入。宿主 wheel 安装、容器启动和编译计数为零。运行前复核镜像摘要、输入哈希、磁盘与标签归属。拟运行上限为一个容器、4 核 CPU 配额、8 GiB 内存、512 个 PID、90 分钟，测试并行度为 4。启动时至少保留 12 GiB 可用盘，降至 8 GiB 以下、超过时限或发生错误时结束本次运行。结束时核对并清理本次标签对应的容器，保留镜像和工作区内的日志及输出。

窗口验收要求原生配方、Pillow 编译、`auditwheel repair` 与三条上游测试全部通过。成功时只接收一个 CPython 3.12 / manylinux_2_28 AArch64 候选 wheel，并记录 SHA-256、动态链接和许可告知。随后在目标 SDK CPython 3.12.16 验证实际加载和受影响流程，再完成阶段 0 的许可交付复评。[主控、导入与运行参数回执](../../scripts/eval/development_probe/sdk_variant/pillow-build-window-preflight-20261119.json)

#### 首次禁网构建停窗

所有者批准一次构建。运行前逐项复核暂存项目、离线输入、宿主工具和固定镜像，通过后启动 cibuildwheel 4.3.1。主控在宿主执行 `CIBW_CACHE_PATH.mkdir` 时试图创建 `~/Library/Caches/cibuildwheel`，被沙盒拒绝，2.075 秒后以状态码 1 退出。此前 `--print-build-identifiers` 在缓存初始化前返回，预检漏掉了这一步。

按失败即停结束本次窗口。Docker 标签查询和事件记录均为空，容器创建和残留各为零。输出目录为空，原生编译、`auditwheel repair`、三条上游测试与产物核对均待实际运行。完整日志及其 SHA-256 已留存。[执行回执](../../scripts/eval/development_probe/sdk_variant/pillow-build-window-result-20261119.json)

后续构建应在启动前将 `CIBW_CACHE_PATH` 指向本次工作区内的独立可写目录。已用同一 cibuildwheel 版本离线验证目录和 `CACHEDIR.TAG` 可创建，验证未调用构建入口。再次提交构建命令须获准开启新窗口。

#### 2026-11-21 禁网重试与停窗

所有者批准一次重试。cibuildwheel 4.3.1 使用工作区本次专用的 `CIBW_CACHE_PATH`，通过缓存初始化后启动唯一容器。容器使用同一固定镜像，实测为 `--network=none`、4 核 CPU 配额、8 GiB 内存和只读的离线输入挂载。Meson 和 Ninja 从本地 wheelhouse 安装，原生构建进入 HarfBuzz 与第二轮 FreeType。

Codex Desktop 更新重启打断宿主构建进程。宿主日志停在 HarfBuzz，恢复取得的容器日志末尾显示第二轮 FreeType 安装和 `before_all` 命令返回 0。Docker 事件记录一次创建、启动、SIGTERM、退出和删除，容器运行约 3 分 27 秒。已按本次标签及完整 ID 清理，复核残留为零。宿主日志、容器日志末尾及事件摘要均已留存。[重试执行回执](../../scripts/eval/development_probe/sdk_variant/pillow-build-window-retry-result-20261121.json)

本次输出目录为空。Pillow wheel 编译、`auditwheel repair`、三条上游测试与产物 SHA-256、动态链接和许可告知核对仍待完成。本次授权的一次重试已结束。下次获准运行时，构建监督与退出记录应可在桌面重启后接续。

#### 2026-11-21 独立监督构建与 wheel 验收

所有者批准一次重试。宿主使用用户域的一次性 `launchd` 任务。探针在启动命令返回后写入记录，实测父进程为 PID 1，`launchd` 保留探针退出码 23。正式任务以同样方式保存监督日志、退出状态与 Docker 生命周期，运行期间未人为重启 Codex Desktop。构建复用固定镜像、只读离线输入及本次专用缓存。唯一容器实测禁网、4 核 CPU 配额和 8 GiB 内存，运行窗口共 211.215 秒，主控退出码为 0。构建容器与一次性监督任务均已清理。

`auditwheel repair` 通过，产出一个 CPython 3.12 AArch64 wheel，包含 `manylinux_2_27_aarch64` 和 `manylinux_2_28_aarch64` 两个标签。上游 `selftest.py` 通过 59 项，`checks/check_wheel.py` 通过 3 项，全量 pytest 通过 5697 项、跳过 119 项，3 项预期失败。wheel 为 22597945 字节，宿主复算与构建日志的 SHA-256 一致，138 条 `RECORD` 核对通过。

核对了 26 个 AArch64 ELF 中的动态链接，包括 8 个扩展、18 个随包共享库和 23 条通过 `$ORIGIN` 解析的包内依赖。另登记 8 个外部 SONAME。候选引用的 108 个 CPython 符号均在 SDK CPython 3.12.16 的已验 `libpython` 中找到，目标 SDK 的实际加载和受影响流程仍待运行。

wheel 内的 `LICENSE` 逐字节匹配 Pillow 正文与 28 段已准备的告知。验收发现 Raqm 已编入扩展，其源码专属版权声明未出现在 wheel 告知中。LibXau 由 `auditwheel` 随 Libxcb 带入，构建器 RPM 声明为 MIT，wheel 中尚未对照到它的专属许可正文。FriBidi 共享库仅供容器测试，成品保留动态加载 shim。后续先闭合具名告知，再完成目标 SDK 运行复验和阶段 0 许可交付复评。[构建、测试、摘要与许可回执](../../scripts/eval/development_probe/sdk_variant/pillow-build-window-launchd-result-20261121.json)

#### 2026-11-21 候选告知与目标 SDK 复验

以构建用 Pillow 13.0.0 源码归档为准，确认 Raqm 0.11.0 的 `COPYING`、版本头内的 Google 许可正文及生成表所用的 Unicode 17.0.0 数据与告知。LibXau 1.0.9-3.el8 的 AlmaLinux AArch64 二进制 RPM 和对应源码 RPM 均通过签名与 payload 摘要验证。两包的 `COPYING` 逐字节一致。二进制 RPM 的 ELF 匹配固定构建器，且 `.text` 与修复后 wheel 中的 LibXau 一致。五份正文与候选 wheel SHA-256 `ff0629b09d29f5cf5f1d560aa0dcdddc64edff0c3df3842bccabb32de26d8460` 绑定，并通过实际 pip 安装报告核验。[许可来源](../../scripts/eval/development_probe/sdk_variant/pillow-notice-sources-20261121.json) · [伴随告知清单](../../scripts/eval/development_probe/sdk_variant/pillow-notice-supplements/manifest.json)

在固定 SDK 镜像的 CPython 3.12.16 / AArch64 中，以 UID 65532 在禁网、只读根容器的 `/tmp` 安装该精确 wheel，`pip check` 通过。实际映射 26 个候选 ELF 和 8 个系统库。7 个扩展完成导入，Tk 扩展因基础镜像缺少 `libtk8.6.so` 维持原有行为。六种图像编解码与 ICC 色彩转换通过。默认环境缺少 FriBidi，与旧 wheel 一致，基础文本渲染通过。单独使用已验 FriBidi 实物后，Raqm 0.11.0 的从右到左文字渲染通过。

SDK 同步与异步请求通过回环固定响应验证，分别将 1 张图缩至 8000 像素长边、21 张图缩至 2000 像素长边。4 项遥测可选依赖检查、7 项 LLM 子合同及 3 项 SDK 工作进程用例通过。前两次探针分别修正了告知文件名和 Tk 可选行为的断言。最终复验退出码为 0，三次探针的所属容器均已清理。最终 SDK 镜像、告知的镜像内交付、伴随源码交付和联合运行仍待单独验收。[SDK 离线执行回执](../../scripts/eval/development_probe/sdk_variant/pillow-sdk-recheck-result-20261121.json)

#### 2026-11-21 最终 SDK 交付离线预检

固定目标为 `linux/arm64`、CPython 3.12.16 和历史验收的 `python:3.12-slim@sha256:aa943ce411423c7582288f18707ff66eddcbaeb510502d747a72aeb34f30d159`。从已验 119 项锁仅替换 Pillow 一行，其他 118 个 wheel 逐字节复用。新锁 SHA-256 为 `c2c214b42b3f443d89854bb24a8dab29c3d2336b2cd4d2d57b401fa8f89304ab`。以目标平台和 Python marker 执行禁网、强制哈希的 pip 试解析，119 项均命中本地 wheel。五份 Pillow 告知已并入交付清单，连同已有 wheel、pip 和 CPython 补充告知共 42 份。[离线预检回执](../../scripts/eval/development_probe/sdk_variant/sdk-delivery-preflight-20261121.json)

伴随源码按新锁绑定。复算了 181 个 SDK 相关留存归档、61 份 Debian `.dsc` 和各自 `Checksums-Sha256` 文件集，另核对 15 个首批覆盖源码映射条目、Pillow 的 31 个源码及配方输入和 SDK 变体的 3 个构建输入。历史归档和取证文件的字节齐备记录沿用已验分片。精确来源与完整覆盖范围仍有具名待办，接收方获取尚待验收。[伴随源码清单](../../scripts/eval/development_probe/sdk_variant/sdk-delivery-sources-20261121.json)

| 验收环节 | 通过条件 |
| --- | --- |
| 禁网构建 | 选定构建器能在本地按完整摘要解析固定 AArch64 基础镜像，禁止访问外部 registry，构建步骤使用 `--network=none`。pip 安装报告须与 119 项锁的名称、版本和本地 wheel SHA-256 全量一致，`pip check` 通过，并记录实际镜像身份。OrbStack 当前查不到该基础镜像引用，构建前须先明确本地供给。 |
| 非 root 告知读取 | 以最终镜像默认身份 `65532:65532`，在禁网、只读根文件系统的容器内执行告知校验。直接读取镜像自带的 19 份 wheel 告知、21 份 pip 告知、2 份 CPython 告知和源码获取说明，核对安装报告与文件摘要。 |
| 伴随源码交付 | 以交付镜像的实际摘要封定 `sources/<SDK-image-sha256>/` 和 `manifest.json`。清单绑定镜像、新锁、组件和相对路径。接收方通过实际交付渠道取得全套对应文件，并复算每项大小及 SHA-256。 |
| SDK 与联合复验 | 用最终镜像摘要复验 Pillow 实际加载、SDK 图像处理、遥测可选依赖、LLM 子合同及 SDK 工作进程。联合用例覆盖正常修改、测试、保存和导出，旧测试证据失效及命令中取消。沿用现有隔离和权限断言，记录容器清理结果。 |

本轮只生成并验证本地构建上下文。实际镜像构建、镜像内告知读取、联合运行和源码交付待后续验收。[交付配方与准备命令](../../scripts/eval/development_probe/sdk_variant/README.md#最终-sdk-交付离线准备)

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
