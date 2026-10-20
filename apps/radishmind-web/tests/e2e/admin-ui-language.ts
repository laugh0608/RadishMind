// Visible product expectations are deliberately independent of runtime dictionaries.
export const providerEn = {
  "action": "Action",
  "activationReason": "Activation reason",
  "candidateId": "Candidate ID",
  "displayName": "Display name",
  "loadCandidate": "Load candidate",
  "operations": {
    "activated": "Activation committed as a new generation.",
    "candidateCreated": "Immutable candidate created from the exact draft revision.",
    "draftSaved": "Draft saved without changing the active Gateway snapshot.",
    "reviewRecorded": "Review recorded. Gateway behavior remains unchanged until activation.",
    "rolledBack": "Rollback committed as a new generation.",
    "workspaceLoaded": "Workspace synchronized. Missing draft or active snapshot remains an explicit empty state."
  },
  "reviewReason": "Review reason",
  "saveDraft": "Save draft revision"
} as const;
export const providerZh = {
  "action": "动作",
  "activationReason": "激活理由",
  "candidateId": "候选 ID",
  "displayName": "显示名称",
  "loadCandidate": "加载候选",
  "operations": {
    "activated": "激活已提交为新代次。",
    "candidateCreated": "已从精确草案修订创建不可变候选。",
    "draftSaved": "草案已保存，当前网关快照保持不变。",
    "reviewRecorded": "审查已记录；激活前网关行为保持不变。",
    "rolledBack": "回滚已提交为新代次。",
    "workspaceLoaded": "工作区已同步；草案或当前快照缺失时明确显示为空。"
  },
  "reviewReason": "审查理由",
  "saveDraft": "保存草案修订"
} as const;
export const quotaEn = {
  "confirmUpdate": "Confirm update",
  "confirmationLabel": "Quota policy update confirmation",
  "missingTitle": "No quota policy exists for this exact application scope",
  "operations": {
    "conflict": "The expected version is stale. Reload the quota owner before reviewing another update."
  },
  "reload": "Reload current policy",
  "requestLimit": "Request limit",
  "reviewUpdate": "Review update"
} as const;
export const quotaZh = {
  "confirmUpdate": "确认更新",
  "confirmationLabel": "配额策略更新确认",
  "missingTitle": "此精确应用范围尚无配额策略",
  "operations": {
    "conflict": "预期版本已过期；请重新加载当前配额策略，再审查更新。"
  },
  "reload": "重新加载当前策略",
  "requestLimit": "请求上限",
  "reviewUpdate": "审查更新"
} as const;
export const pricingEn = {
  "confirm": "Confirm future pricing",
  "inputRate": "Input micro-USD / 1M",
  "loadOwner": "Load exact pricing owner",
  "missingTitle": "No policy exists for this exact scope",
  "modelId": "Model ID",
  "outputRate": "Output micro-USD / 1M",
  "profileId": "Profile ID",
  "providerId": "Provider ID",
  "reload": "Reload current revision",
  "reviewRevision": "Review revision",
  "sanitizedReason": "Sanitized reason",
  "scopeChanged": "Scope changed. Load the exact owner before editing; the previous revision is no longer actionable."
} as const;
export const pricingZh = {
  "confirm": "确认后续请求价格",
  "inputRate": "输入微美元 / 百万 token",
  "loadOwner": "加载精确价格策略",
  "missingTitle": "此精确范围尚无价格策略",
  "modelId": "模型 ID",
  "outputRate": "输出微美元 / 百万 token",
  "profileId": "配置档案 ID",
  "providerId": "模型提供方 ID",
  "reload": "重新加载当前修订",
  "reviewRevision": "审查修订",
  "sanitizedReason": "脱敏理由",
  "scopeChanged": "范围已改变，请先加载精确价格策略再编辑；此前修订不能继续操作。"
} as const;
export const readEn = {
  "auditTitle": "Sanitized audit window",
  "deleteBlocked": "Delete blocked",
  "tenantBoundary": "This is a sanitized summary projection. It cannot create a tenant, edit membership, change plan or quota, reveal a raw tenant record, or establish production authorization."
} as const;
export const readZh = {
  "auditTitle": "脱敏审计窗口",
  "deleteBlocked": "不可删除",
  "tenantBoundary": "此处仅展示脱敏摘要，不能创建租户、编辑成员、修改方案或配额、暴露原始租户记录，也不构成生产授权。"
} as const;
