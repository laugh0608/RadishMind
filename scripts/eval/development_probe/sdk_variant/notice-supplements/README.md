# SDK 候选的补充告知

本目录保存从已取得材料复制的许可正文或完整许可注释块，适用于 [license-delivery-20261105.json](../license-delivery-20261105.json) 中绑定的精确 wheel。原文内容保持英文；fastmcp-slim 原文没有末尾换行，为满足仓库文本规范仅补一个 LF，原始与输出摘要分别登记。其余文件逐字节保留；来源归档、成员路径和注释块字节范围均在清单登记。

| 对象 | 本目录补充内容 | 尚未证明的范围 |
| --- | --- | --- |
| fastmcp-slim 3.4.11 | 精确上游提交的 Apache 正文；源码归档与 wheel 的 270 个包内文件相同 | 完整上游树没有 NOTICE 命名文件、LICENSE blob 已核对；实际镜像告知安装待完成 |
| cryptography 内嵌 OpenSSL 4.0.3 | 精确上游版本的 Apache 正文，版本与 wheel 内 SBOM 一致 | 对应源码归档字节、完整构建来源与实际镜像安装 |
| tokenizers 0.23.3 | 精确源码归档内的 Apache 正文 | Rust 实际链接组件及其告知、实际镜像安装 |
| Lupa 2.8 | Lua 5.1 至 5.5 的许可块、Lua 5.1 的 COPYRIGHT、LuaJIT 2.1 的 COPYRIGHT 与 Lua 许可块 | 原生构建来源完整映射、实际镜像安装 |

本目录是补充材料，不是完整第三方告知集合，也不替代 wheel 已有的许可证。不得用这些文件覆盖原告知，或把源码树中的 LuaJIT 2.0 计作已交付组件。

候选镜像内目标目录为 `/usr/local/share/radishmind/third-party/`，逐文件位置见清单；本轮没有修改镜像或构建配方，`installed_in_image` 均为 `false`。后续镜像构建范围确认后，须连同既有告知、完整组件清单、覆盖源码与获取说明一起纳入交付，并检查实际镜像中的字节。OpenSSL 4.0.3 对应源码取得与独立构建核验、regex 历史条款、其他原生组件和覆盖源码交付仍未闭合，不能据此宣布整镜像可分发。
