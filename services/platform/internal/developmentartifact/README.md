# 开发产物负责模块

固定 Go 探针的显式保存、精确读取与即时 JSON 导出。完整合同见[开发产物合同](../../../../docs/contracts/development-artifact-v1.md)。

- `Build / Parse` 生成和严格校验开发包；基准、候选、patch 与测试证据绑定，保留缺失 / 过期 / 未知事实。
- `Service.Preview / Save / Read / Export` 分别处理预览、摘要确认、精确读取和独立导出权限；包不可变，预览不写资产表。
- `SQLiteStore` 使用共享数据库的独立 `development_artifacts` 迁移；事务检查跨应用的所有者配额，不回退内存。
- 可信宿主通过 `developmentsandbox.NewDeliverySource(runtime, session, identity)` 提供关闭后的来源，再使用 `Server.ConnectDevelopmentArtifacts(verifiedRequest, source)` 创建请求绑定服务。正式 HTTP、Web 与内置角色自动授权未开放。
- 临时材料创建时固定 24 小时期，保存失败不续期；已保存包不自动过期。读取 / 导出不依赖模型或执行器。

定向验证不需要 Docker、服务或模型：

```bash
cd services/platform
go test -race ./internal/developmentartifact ./internal/developmentsandbox ./internal/strictjson ./internal/workspacepolicy
go test ./internal/httpapi -run TestDevelopmentArtifact
```

上述定向测试使用合成源码、隔离临时 SQLite 和受控身份对象，不调用真实执行入口。另行授权的[真实 SDK / 代码沙箱联合验收](../../../../docs/devlogs/2026-W44.parts/development-joint-20261031.md)已覆盖同次任务保存 / 导出、候选变化后旧测试失效及命令中取消；[遥测可选化候选](../../../../scripts/eval/development_probe/sdk_variant/README.md)也已独立复验。两轮运行窗口均已结束，不构成正式 HTTP、Web 或真实模型效果验收。
