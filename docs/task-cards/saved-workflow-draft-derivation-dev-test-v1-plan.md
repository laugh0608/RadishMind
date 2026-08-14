# 已保存 Workflow 草案派生（开发 / 测试态）v1 实施任务卡

更新时间：2026-07-27

状态：`completed`

## 目标

交付从精确已保存草案版本派生独立本地草案的完整开发 / 测试态路径，并让脱敏直接来源随既有 Saved Draft 保存 / 恢复链往返。

## 范围

1. 在 Web 建立可单测的纯派生 builder，深复制图、布局和审查字段，生成稳定短 `draft_id`。
2. Draft Designer 只在活动内容与精确 saved version 一致时开放派生。
3. 新草案以版本 `0` 进入既有 Validate / Save / Review，不修改源记录。
4. 通过 `additional_fields.derivation_v1` 保存直接父草案 ID 和版本。
5. Go domain 对 `derivation_v1` 做结构归一化，并拒绝自引用。
6. 同步 Workflow 入口、当前焦点、能力矩阵、路线图和本周周志。

## 验证

```bash
(cd services/platform && go test ./internal/httpapi)
npm --prefix apps/radishmind-web test
npm --prefix apps/radishmind-web run build
git diff --check
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```

## 停止线

- 不新增后端 route、repository、migration、owner、发布记录或运行记录。
- 不实现自动保存、覆盖、合并、同步、跨 application / workspace 派生。
- 不复制敏感输入输出、身份、membership、凭据、完整审计记录或 provider payload。
- 不新增 fixture / checker；相邻 Go / Web 测试和既有聚合门禁负责验证。
