# `Radish` 任务卡：`answer_docs_question`

更新时间：2026-03-31

## 任务目标

基于 `Radish` 的固定文档、在线文档和当前页面上下文，输出可引用、可追溯、适合直接展示的文档问答结果。

当前任务是 `Radish` 的首个最小接入场景，用来验证统一协议在“知识优先、低风险、可评测”路径上的可行性。

## 请求映射

- `project`: `radish`
- `task`: `answer_docs_question`
- 请求结构遵循 [CopilotRequest schema](../../contracts/copilot-request.schema.json)

## 最小必需输入

请求至少应包含以下内容：

- `context.current_app`
- `context.route`
- `context.resource`
- 与当前问题直接相关的文档型 `artifact`

推荐至少提供一个主输入文档 artifact：

- `kind`: `markdown` 或 `text`
- `role`: `primary`
- `name`: 例如 `wiki_document`、`docs_page`

如果调用方已经做过检索，也可以补充 supporting/reference artifacts，但不应把无关长文整包透传。

## 可选补充输入

- `context.viewer`
- `context.search_scope`
- `context.attachment_refs`
- 已召回的文档片段、FAQ 片段或论坛摘录

这些输入用于增强答案质量，但不应替代当前页面或当前问题本身。

## 检索边界

当前任务的检索目标是“帮助回答当前文档问题”，不是把 `Radish` 的全部知识源混成一个泛化助手入口。

当前推荐的知识源优先级如下：

1. 当前页面直接对应的固定文档或在线文档
2. 与当前 `resource` 同主题的 docs/wiki 片段
3. 明确被当前页面或文档引用的附件说明
4. 仅在前述证据不足时，补充受控的论坛摘录或 FAQ 片段

当前不建议优先检索或透传的内容：

- 与当前问题无关的论坛长线程
- 需要最终权限判定的 Console 内部状态
- 未脱敏的后台运维记录或用户私密内容
- 只与运营经验相关、但没有正式文档支撑的口径

如果正式文档与论坛经验口径冲突，应优先以正式文档为主，并在响应里显式说明来源差异，而不是静默混合。

## 召回输入规范

当调用方已经完成检索时，推荐把召回结果压成“小而可引用”的输入包，而不是把整篇长文直接塞进请求。

当前建议的召回输入组织方式：

- 1 个 `primary` artifact：当前页面或当前问题最直接相关的主文档
- 0 到 3 个 `supporting` artifacts：补充片段、相关页面或 FAQ 摘录
- 0 到 2 个 `reference` artifacts：仅用于解释来源差异或导航补充

每个召回 artifact 建议满足以下约束：

- 单个 artifact 优先是片段级，而不是整站级汇总
- `name` 能反映来源类型，例如 `wiki_document_chunk`、`docs_page_excerpt`
- `metadata` 至少应能标识来源页面、片段序号或召回分数
- 若来自论坛或 FAQ，必须让调用方能够在 `label` 或 `metadata` 中区分其不是正式文档

当前建议的片段粒度：

- 优先 1 到 3 个自然段
- 或 1 个完整的小节
- 避免把超过当前问题所需范围的上下文一并召回

如果当前页面已经能直接回答问题，则不应为了“多给模型一点背景”而继续扩张召回范围。

## 禁止透传

以下内容不应进入本任务请求：

- `access_token`、`refresh_token`、cookie
- 原始 OIDC 响应、临时访问令牌或安全凭据
- 与当前问题无关的大段 Console 内部状态
- 未经裁剪的附件原始私密内容

## 输出要求

响应结构遵循 [CopilotResponse schema](../../contracts/copilot-response.schema.json)。

本任务对字段的要求如下：

- `summary` 必须先直接回答问题，而不是先铺背景
- `answers` 至少包含一个可展示回答片段
- `issues` 仅在证据不足、口径冲突或无法确认时使用，不要把正常问答写成问题列表
- `citations` 必须引用具体文档、页面或片段
- `proposed_actions` 默认为空；若存在，只能是低风险的导航或补充阅读建议

## 风险分级规则

- `low`: 纯文档问答、引用说明、页面导航建议
- `medium`: 涉及权限理解、操作前提或流程说明，但仍停留在解释层
- `high`: 试图替代权限最终判定、授权决策、治理结论或直接给出高风险操作指令

## `requires_confirmation` 触发条件

通常情况下本任务应为 `false`。

仅当出现以下情况时应为 `true`：

- 输出包含可能改变系统状态的候选动作
- 输出涉及权限授予、治理结论或高风险运营操作建议
- 输出证据不足，但仍需操作者据此采取敏感动作

## 正反例边界

正例：

- 回答某篇文档里明确写明的能力说明
- 结合当前页面和文档片段解释一个功能入口在哪里
- 证据不足时明确说“不确定”并指出应查看的文档来源

反例：

- 把文档中没有明确说明的权限结论说成既定事实
- 用论坛经验帖替代正式文档口径，且不标明来源差异
- 直接输出“已为你授权/已替你操作完成”之类越界内容

## 最小输入快照示例

```json
{
  "project": "radish",
  "task": "answer_docs_question",
  "artifacts": [
    {
      "kind": "markdown",
      "role": "primary",
      "name": "wiki_document",
      "mime_type": "text/markdown",
      "uri": "wiki://guide/authentication"
    }
  ],
  "context": {
    "current_app": "document",
    "route": "/document/guide/authentication",
    "resource": {
      "kind": "wiki_document",
      "slug": "guide/authentication",
      "title": "鉴权与授权指南"
    },
    "search_scope": [
      "docs",
      "wiki"
    ]
  }
}
```

## 评测口径

当前阶段至少检查以下指标：

- 结构合法率：响应必须通过 schema 校验
- 直接回答率：`summary` 和主要 `answers` 必须先回答问题，而不是只复述背景
- 引用命中率：关键结论必须能对上具体文档或片段
- 证据克制度：证据不足时应显式降级，而不是编造确定性答案
- 风险分级正确率：普通文档问答应稳定落在低风险

## 非目标

当前任务不负责：

- 替代 `Radish` 的权限最终判定
- 直接执行 Console、治理或运营动作
- 先做泛化助手式多工具自治
