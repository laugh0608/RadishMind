# 2026-10-31 SDK、代码沙箱与产物联合验收

## 授权与范围

起点为干净 `dev@eb1b2ef9`。项目所有者要求阶段 0 的 SDK 与代码沙箱同次任务联合验收，覆盖候选修改、测试证据和产物保存 / 导出；随后明确批准可信控制容器的 Docker 管理通道、Linux Docker CLI 准备及隔离运行窗口。未运行真实模型、正式 Web、固定团队或用户试用；未推送。

复用 SDK 1.59.0 镜像 `sha256:d80a7496e7c659f1fc76ffb5f429530bcfcd1fef5a790fa06b789ff27eb23842` 与 Go 1.26.7 镜像 `sha256:b17af760035fc2f338eed92d448a6c67f2d45438844fc6c60678fa5f99e44b57`，没有重建或下载镜像。仅从 Docker 官方静态分发下载 `linux/aarch64` CLI 29.8.2 归档，并提取 `docker/docker` 到忽略目录；归档 SHA-256 为 `76a624e4a8e5da654d1150e808175125efb5a6f1b6aa1cbd9caee18f51047a50`，CLI 为 `37a53fed3445028a49443fd6f9338ccedeeeda9776de5d7fd4d7e7c936b789cc`。未安装系统依赖或修改 Python 锁。

## 实现与边界修正

新增 `developmentsandbox` 的可选联合 Go 测试与 `scripts/eval/development_probe/run_joint.py / joint_worker.py`。复用既有 `RunWorker → WorkerLaunch.Sandbox → NewDeliverySource → developmentartifact.Service`；没有新增生产 API、schema、存储组件或自动授权。

Go 控制程序在专用 root 容器中管理 Docker，SDK 启动器在读取任何任务凭据前降至 `65532:65532`，清空附加组，并核对有效 / 许可 / 继承 / ambient capability 为零及禁止提权。控制程序仅保留 `CHOWN / SETUID / SETGID / DAC_OVERRIDE / KILL`；代码容器仍使用既有非 root、无网络、只读源码及独立 watchdog 档案。

首次启动边界检查拒绝继续；第二轮定位为本机共享挂载没有按显示的 `0700` 阻止 SDK 读取控制数据库，实际打开文件也成功，不能仅以模式位声明隔离。改为在容器内用 Linux 私有 tmpfs 遮蔽材料挂载点的父目录；管理 socket 也放在另一处 Linux 私有目录内。最终 SDK 实际打开数据库及连接 Docker socket 均得到权限拒绝，保存 `sdk-boundaries.json`。未放宽断言，也未在 SDK 中开放文件 / 终端工具。

控制容器只读根、无外网、无宿主端口，1 GiB 内存且无额外 swap、2 CPU、128 进程、512 MiB 临时区，两个私有目录各 1 MiB；只挂载本次材料、明确的程序 / SDK 适配目录及已授权管理 socket。服务仅在同容器回环临时端口监听。实际配置先 inspect 核对再启动。

第三轮三条业务链已取得产物，但测试清理重复关闭已释放数据库上的 Runtime 导致整轮失败；修正测试关闭顺序后第四轮全部通过。随后补充失败路径清理顺序与宿主精确资源核对；第五轮启动前检查发现 Docker inspect 返回 `CAP_*` 规范名，修正期望后，第六轮三项全部通过。前述失败证据均保留，没有将部分结果算作整轮通过。

## 最终联合证据

每条用例使用新的尝试、SQLite 和候选副本。提供方固定响应及资格 / 产物权限适配为合成实现；真实 SDK `Agent / LocalConversation`、内部通道、持久控制、文件工具、代码容器、可信来源与产物存储均走既有实现。本轮不证明正式身份入口或真实提供方原生传输。

| 用例 | 实际观察 |
| --- | --- |
| 正常交付 | SDK 经三次工具调用读取 `list.go`、替换候选、执行固定测试；4 次模型请求、120 个合成实报 token；候选的 8 项冻结子用例通过，基准未改变；包状态为 `command_passed` |
| 旧测试失效 | SDK 先读取、执行基准测试，观察两项预期失败，再替换候选；4 次模型请求、120 token；包保留原测试摘要、失败日志和 `stale`，不显示当前候选通过 |
| 命令中取消 | SDK 写入只用于取消的持续运行候选；实际代码容器运行后提交持久 CAS 取消；3 次模型请求、90 token；从取消提交到 SDK 退出 / 空组及沙箱关闭返回约 120 ms，低于本例 10 秒断言；包保留 `unknown`，不伪造测试通过或完成 |
| 三例共有保存链 | 预览不写表，错误摘要拒绝；注入 SQLite 保存失败无部分写入；撤销故障后相同请求可保存且幂等；关闭 Runtime / 数据库后重开并精确读取尝试和包，不重放模型 / 工具；导出严格解析及摘要重验；撤销 export 权限后导出拒绝、独立 read 仍可用 |

保存 / 导出均保留 `not_reviewed / not_accepted`。固定响应由测试程序明确选择修改内容，不能称为模型自主修复、真实开发质量或人工接受。取消用例证明本地这一受控链路，不推定真实上游停止计算或零费用。

证据位于忽略目录 `tmp/development-joint-20261031/`：最终 `run-06.log`、命令及程序摘要、控制容器实际配置 / 退出状态和 `.cleanup.json`；各 case 保存事件、边界报告、取消耗时、资源账册、容器配置 / 退出记录、验证报告、SQLite 及 `export.json`。每次最多 3 次工具和 4 次模型请求，未修改冻结任务文件、测试断言或预算。

## 清理、回归与下一步

最终运行退出 0、无 OOM；逐轮移除 6 个精确控制容器。`cleanup-audit.json` 对全部 15 份归属账册查询，涉及的 9 个代码容器均无残留；最终三份账册均无需宿主补救删除。所属 SDK 进程、回环监听器随控制容器退出，未留下后台服务。保留已有镜像、CLI 和隔离审计材料，不清理其他任务资源。

全平台 `GOPROXY=off go test -race -timeout 20m ./...` 与 `GOPROXY=off go vet ./...` 通过，日志为 `go-race-all.log / go-vet.log`；最终测试关闭顺序变更后，相关三包定向 race 再过，见 `go-final-focused.log`。Python development 聚合 19 项通过（`python-development.log`），最终三份真实导出包另通过 canonical schema 验证（`export-schema-audit.json`）。`./scripts/check-repo.sh --skip-platform-go-tests` 全量非 Go 门禁退出 0（`check-repo-full.log`），Go 已由上述全平台 race / vet 单独覆盖。最终文字再核对篇幅、链接和 `git diff --check`；既有篇幅提示保留，W44 新增记录置于分片，周志入口仍在原篇幅内。重复 `run-id` 的准备阶段拒绝检查也通过，不覆盖失败证据。普通 Go 环境联合测试明确 skip，只有上述授权镜像中的运行计为联合证据。

已同步主 / 子专题、集中任务卡、产物 / 尝试 / 沙箱合同、当前焦点、路线图、能力矩阵、架构、集成契约和复验说明。下一步按阶段 0 关键条件进行采用复评，收口依赖许可证、提供方原生传输与阶段 1 的 UI / 实现范围。SDK 保持首选候选，阶段 0 尚未退出；本次技术证据不自动启动正式 Web、通用仓库、真实模型或固定团队。
