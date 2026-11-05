# SDK 候选的补充告知

本目录保存从已取得材料复制的许可正文或完整许可注释块，适用于 [license-delivery-20261105.json](../license-delivery-20261105.json) 中绑定的精确 wheel。原文内容保持英文；fastmcp-slim 原文没有末尾换行，为满足仓库文本规范仅补一个 LF，原始与输出摘要分别登记。regex 的 CNRI 最终 1.6 正文从官方页面解码 HTML 并保留全部八条，页面 / 片段 / 输出摘要及规范化步骤见[核对记录](../license-page-result-20261105.json)；未替换其许可用词。其余文件逐字节保留；来源归档、成员路径和注释块字节范围均在清单登记。

| 对象 | 本目录补充内容 | 尚未证明的范围 |
| --- | --- | --- |
| fastmcp-slim 3.4.11 | 精确上游提交的 Apache 正文；源码归档与 wheel 的 270 个包内文件相同 | 完整上游树没有 NOTICE 命名文件、LICENSE blob 已核对；实际镜像告知安装待完成 |
| cryptography 内嵌 OpenSSL 4.0.3 | 精确上游版本的 Apache 正文，版本与 wheel 内 SBOM 一致 | 对应源码归档摘要及许可正文已核对；独立构建来源与实际镜像安装仍待验证 |
| tokenizers 0.23.3 | 精确源码归档内的 Apache 正文 | Rust 实际链接组件及其告知、实际镜像安装 |
| regex 2026.9.29 | 官方页面所载 CNRI Python 1.6 最终版全文，与其发行物声明所指版本对应 | CPython 2.6 / 3.1 继承来源与完整告知范围、实际镜像安装仍待核对 |
| Lupa 2.8 | Lua 5.1 至 5.5 的许可块、Lua 5.1 的 COPYRIGHT、LuaJIT 2.1 的 COPYRIGHT 与 Lua 许可块 | 原生构建来源完整映射、实际镜像安装 |

本目录是补充材料，不是完整第三方告知集合，也不替代 wheel 已有的许可证。不得用这些文件覆盖原告知，或把源码树中的 LuaJIT 2.0 计作已交付组件。

候选镜像内目标目录为 `/usr/local/share/radishmind/third-party/`，逐文件位置见清单。离线准备器已按冻结 wheel、正文大小 / 摘要和目标路径复制材料；构建配方已加入复制与安装报告匹配检查。实际镜像尚未重建，`installed_in_image` 均为 `false`；历史镜像不因配方更新而视为已有告知。后续获准构建时仍须检查实际镜像中的字节和非 root 可读性。OpenSSL 4.0.3 的[归档字节已匹配 wheel SBOM](../license-source-result-20261105.json)，独立构建核验、regex 继承来源及告知范围、其他原生组件和覆盖源码交付仍未闭合，不能据此宣布整镜像可分发。
