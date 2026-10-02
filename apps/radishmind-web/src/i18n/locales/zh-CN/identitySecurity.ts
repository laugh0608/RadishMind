export const identitySecurity = {
  "eyebrow": "账户安全 · 开发 / 测试态",
  "title": "会话与本地凭据",
  "claim": "认领邀请",
  "opening": "正在打开…",
  "link": "关联 Radish 身份",
  "signingOut": "正在退出…",
  "signOut": "退出登录",
  "close": "关闭账户安全",
  "currentMethod": "当前登录方式",
  "sessionOwner": "会话标识",
  "recentAuthentication": "近期认证",
  "verified": "已验证",
  "required": "需重新验证",
  "snapshot": "目录快照",
  "pending": "等待中",
  "unknown": "未知",
  "committed": "安全变更已确认",
  "committing": "正在提交安全变更…",
  "owner": "当前账户的会话",
  "directory": "会话目录",
  "loadedMore": "已加载 {{countText}} 条 · 还有更多",
  "loadedComplete": "已加载 {{countText}} 条 · 快照完整",
  "loadingTitle": "正在读取会话",
  "loadingDescription": "正在等待当前账户的会话目录。",
  "emptyTitle": "未返回会话",
  "emptyDescription": "你仍处于登录状态，但服务未返回会话记录。请重新加载目录后再执行安全操作。",
  "currentSession": "当前会话",
  "outsideWindow": "当前会话不在已加载页面中。请加载剩余会话后再进行变更。",
  "otherSessions": "其他有效会话",
  "noOtherSessions": "此快照中没有其他有效会话。",
  "endedHistory": "已结束的会话",
  "noEndedSessions": "此快照中没有已过期或已撤销的会话。",
  "loadingMore": "正在加载下一页快照…",
  "loadMore": "加载剩余会话",
  "invalidTitle": "会话目录无效",
  "invalidDescription": "会话数据与当前会话不一致。请重新加载目录。",
  "actions": "凭据与批量会话操作",
  "credentialAction": "本地密码安全",
  "rotate": "轮换本地凭据",
  "available": "可用",
  "unavailable": "不可用",
  "rotationDescription": "在同一事务中更换本地密码，并撤销使用该凭据创建的全部有效会话。OIDC 会话保持有效。",
  "currentPassword": "当前密码",
  "newPassword": "新密码",
  "passwordHelp": "12–1024 个字符；密码策略由服务端校验。",
  "confirmPassword": "确认新密码",
  "impactConsent": "我已知晓：使用此凭据创建的会话将被撤销。",
  "loadBeforeRotation": "请加载完整目录，审查密码变更将撤销哪些会话。",
  "reviewRotation": "审查凭据轮换",
  "keepCurrent": "保留当前会话",
  "revokeOthers": "撤销其他有效会话",
  "bulkDescription": "一并退出其他所有有效会话，当前浏览器会话保持有效。",
  "loadBeforeBulk": "请加载完整目录后再审查目标会话。",
  "reviewOthers": "审查撤销其他会话",
  "privacyBoundary": "设备指纹、IP 地址、原始 User-Agent、上游 token、密码和会话目录数据均不写入浏览器存储。",
  "thisSession": "当前浏览器会话",
  "lastVerified": "最近验证",
  "reviewRevoke": "审查撤销",
  "confirmRotation": "轮换凭据并撤销其会话？",
  "confirmExact": "撤销此会话？",
  "confirmBulk": "撤销其他所有有效会话？",
  "confirmation": "安全操作确认",
  "currentWillClose": "撤销范围包含当前会话。成功后将清除其认证 Cookie，你需要重新登录。",
  "currentWillRemain": "当前会话保持有效。不在此列表中的 OIDC 会话也保持有效。",
  "targetSet": "已审查的目标会话 · {{countText}}",
  "noTargets": "此快照中没有有效的目标会话。",
  "atomicity": "如果密码更换或任一会话撤销失败，原凭据和所有会话均保持不变。",
  "cancel": "取消并清除输入",
  "rotateAndRevoke": "轮换并撤销",
  "revokeExact": "撤销此会话",
  "commitBulk": "撤销其他会话",
  "reload": "重新加载会话目录",
  "methods": {
    "local_password": "本地密码",
    "oidc": "Radish OIDC"
  },
  "states": {
    "active": "有效",
    "expired": "已过期",
    "revoked": "已撤销"
  },
  "inputErrors": {
    "unavailable": "此账户没有可轮换的有效本地凭据。",
    "recent": "请退出并重新认证，再轮换本地凭据。",
    "incomplete": "请加载完整目录后再审查受影响的会话。",
    "range": "请输入当前密码，以及 12–1024 个字符的新密码。",
    "mismatch": "两次输入的新密码不一致。",
    "impact": "请确认会话影响后再继续。"
  },
  "success": {
    "currentRevoked": "当前会话已撤销。请重新登录后继续。",
    "exactRevoked": "所选会话已撤销。",
    "othersRevoked_one": "已撤销 {{countText}} 个其他有效会话。",
    "othersRevoked_other": "已撤销 {{countText}} 个其他有效会话。",
    "credentialClosed": "已按 {{policyVersion}} 轮换凭据，当前本地会话已结束。",
    "credentialRevoked_one": "凭据已轮换；已撤销使用该凭据创建的 {{countText}} 个会话。",
    "credentialRevoked_other": "凭据已轮换；已撤销使用该凭据创建的 {{countText}} 个会话。"
  },
  "failures": {
    "authentication_required": {
      "title": "需要认证",
      "message": "当前会话已无法授权此操作。请重新登录。"
    },
    "denied": {
      "title": "会话作用域校验未通过",
      "message": "账户、会话、Origin 或 CSRF 校验不再匹配。未应用任何变更。"
    },
    "recent_authentication": {
      "title": "需要近期认证",
      "message": "请退出并重新认证，再重试此安全操作。"
    },
    "conflict": {
      "title": "会话或凭据已变更",
      "message": "提交前会话或凭据发生了变化。请重新加载后再做决定。"
    },
    "credential_unavailable": {
      "title": "本地凭据不可用",
      "message": "此账户没有有效本地凭据。未创建新的登录方式。"
    },
    "credential_invalid": {
      "title": "密码验证未通过",
      "message": "当前密码无效，或新密码与原密码相同。原有状态均保持不变。"
    },
    "credential_policy": {
      "title": "密码策略校验未通过",
      "message": "新密码不符合服务端密码策略。请重新输入当前密码和新密码。"
    },
    "unavailable": {
      "title": "安全服务不可用",
      "message": "服务未能完成此操作。请重新加载并核对当前状态后再重试。"
    },
    "invalid_response": {
      "title": "安全响应无效",
      "message": "响应不符合预期格式或隐私边界，已被拒绝。"
    },
    "failed": {
      "title": "安全请求失败",
      "message": "无法验证操作结果。请重新加载并核对当前状态。"
    },
    "selectionStale": {
      "title": "所选会话已变更",
      "message": "此快照中的目标会话已失效。请重新加载后再做决定。"
    },
    "reviewStale": {
      "title": "凭据审查已失效",
      "message": "提交前凭据输入已清除。请重新输入当前密码和新密码。"
    }
  }
} as const;
