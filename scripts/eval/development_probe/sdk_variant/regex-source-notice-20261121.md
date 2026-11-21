# regex 2026.9.29 伴随源码告知

本告知对应 SDK 镜像清单 `sha256:5e8c4c4473d5497febada6cb563b6630c495b1dcd54651de7c270df38bdd1daf` 中的 Linux AArch64 wheel `regex-2026.9.29-cp312-cp312-manylinux2014_aarch64.manylinux_2_17_aarch64.manylinux_2_28_aarch64.whl`，SHA-256 为 `f93bc1c3486ef3747e07c9d7c1d0a147b8fbaab975f80e348aed6f71309dfaca`。对应发行源码为 `regex-2026.9.29.tar.gz`，SHA-256 为 `8b5fcc4771732191b2b7d1dd68d8f0353f47f8d90b6150f6dce58bf1112442cb`。

## SRE 继承和改动摘要

上游 `LICENSE.txt` 声明 `regex` 继承 CPython 2.6 和 3.1 的 `re`，将继承部分归于 Secret Labs AB 和 CNRI Python 1.6，并以 Apache 2.0 许可其新增和改动。已安装的 `LICENSE.txt` 与发行源码中的正文逐字节相同。发行源码 `src/_regex.c` 和 `regex/_main.py` 分别与两版 CPython 的 `Modules/_sre.c` 和 `Lib/re.py` 保有共同的 Secret Labs 和 CNRI 告知。`src/_regex.c` 还保有与 CPython 3.1 共同的 `Pattern`、`Match`、`Scanner` 初始化代码。`src/_regex.h` 与 `regex/_regex_core.py` 保有各自的上游版权告知。

按 `regex` 发行源码的改动记录，Matthew Barnett 于 2009 年重新设计匹配器，2010 年重写和扩展 Python 前端，并重写 C 引擎头文件，2011 年为引擎加入 PEP 393 字符串支持。RadishMind 采用该原样发行物。CPython 2.6 与 3.1 的顶层 `LICENSE` 全文随附，保留其中的 PSF、CNRI Python 1.6.1 和其他历史告知。继承代码所指的 CNRI Python 1.6 最终版全文另行保留。

## Unicode 生成表

`src/_regex_unicode.h` 与 `tools/build_regex_unicode.py` 声明 Unicode 18.0.0。生成工具以 Unicode Character Database 的 `UCD.zip` 为输入。`setup.py` 将生成的 `src/_regex_unicode.c` 编入 `regex._regex`。对应的 Unicode, Inc. 版权及许可全文由随附的 `Unicode-LICENSE.txt` 提供，适用于该组件的 Unicode 数据生成表。

## 同交付告知文件

这些文件与本告知及精确发行源码登记在同一镜像摘要目录的 `manifest.json`。路径相对于 `python/regex/2026.9.29/`。

| 文件 | 来源 | SHA-256 |
| --- | --- | --- |
| `notices/LICENSE.txt` | 精确发行源码及 wheel 自带告知 | `bff55ef4cdcc8c14ce259f8e8ab60e264418440d6335f4dc138273fbd506144d` |
| `notices/CNRI-Python-1.6-LICENSE.txt` | CNRI Python 1.6 最终版八条正文 | `d4c9c432696dec0f8af87cf38160a671648c3591459d9edc785796938bba7c00` |
| `notices/CPython-2.6-LICENSE.txt` | CPython `v2.6` 顶层 `LICENSE` | `f7a0d9a43d04eade9ecb2f677a617b7c22794355ac49f0b141ab194fb030a64f` |
| `notices/CPython-3.1-LICENSE.txt` | CPython `v3.1` 顶层 `LICENSE` | `1524dab343bfa65c65f0450cebaf6472d98f668ecf615df1eadb149f88912fff` |
| `notices/Unicode-LICENSE.txt` | 已留存的 Unicode, Inc. 通用 Unicode License V3 全文 | `e7a93b009565cfce55919a381437ac4db883e9da2126fa28b91d12732bc53d96` |
