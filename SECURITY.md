# RadishMind 安全策略

RadishMind 当前处于内部开发者预览阶段，尚未承诺长期安全支持的稳定版本或固定响应 SLA。但项目涉及账户与联合身份、工作区授权、API 密钥、模型与 Provider 调用、受控工具执行、运行证据和用户输入，安全问题应优先私下处理。

## 私下报告

请不要为未修复漏洞创建公开 Issue、Pull Request 或讨论，也不要在报告中附带真实密钥、访问令牌、个人数据、生产凭据或未经授权取得的第三方内容。

本仓库已启用 [GitHub Private vulnerability reporting](https://github.com/laugh0608/RadishMind/security/advisories/new)，请优先通过该私密入口报告漏洞。若该入口暂时不可用，可发送邮件至 `laugh0608@foxmail.com`，主题包含 `[RadishMind Security]`。

报告应尽量包含：

- 受影响的提交、版本、组件、部署模式或存储后端；
- 可复现步骤、最小输入、预期行为和实际行为；
- 对机密性、完整性、可用性、隐私或授权边界的影响；
- 已脱敏的日志、请求、响应、运行记录或证据；
- 已知利用条件、临时缓解措施和建议披露时间；
- 是否可以公开致谢及希望使用的署名。

若复现需要敏感材料，请先说明材料类型并等待安全传输安排，不要直接通过普通 Issue、PR 或邮件正文发送秘密。

## 安全问题范围

- 本地账户、Session、Radish OIDC、Authorization Code + PKCE、CSRF、Origin 或 Cookie 策略可被绕过；
- workspace membership、角色、权限、application scope 或资源 owner 校验缺失，导致越权读取、写入、执行或跨作用域数据泄露；
- API Key、Provider 凭据、端点、用户输入、模型提示、评测数据或结果资产通过日志、错误、前端持久化、导出或运行记录泄露；
- 模型输出、Prompt Injection 或恶意外部内容能够绕过结构化协议、人工确认或规则复核，触发未授权工具调用、业务写回或高风险动作；
- 受控 HTTP Tool、Provider 调用、文件处理或导入链路存在 SSRF、路径穿越、任意代码执行、请求走私、资源失控或拒绝服务风险；
- memory、SQLite、PostgreSQL 或外部 Provider 之间发生未声明 fallback、作用域漂移、版本校验绕过或失败后副作用；
- audit、manifest、digest、evaluation、artifact、candidate record 或其它证据可被伪造、替换、截断或错误归属；
- 依赖、构建、CI、发布、模型、数据集或第三方资产存在供应链完整性风险。

普通功能缺陷、模型质量问题、文档错误和不产生安全边界影响的性能问题可以使用公开 Issue。模型给出不准确建议本身不自动构成漏洞；若该输出能够绕过授权、确认、协议或执行停止线，则按安全问题处理。

## 处理与披露

当前阶段按 best-effort 方式确认、评估和修复报告，尚不承诺固定响应时间、受支持版本范围或公开发布时间。报告者与维护者应协调披露，在修复或有效缓解措施可用前避免公开可直接利用的细节。

安全修复仍通过受保护分支和 Pull Request 流程进入稳定主线；必要时可使用私有修复分支或 GitHub Security Advisory 的临时私有分叉。直接进入 `master` 的 hotfix 合并后必须回同步到 `dev`，且不得把真实秘密、个人数据或可利用载荷写入长期 fixture、日志或文档。
