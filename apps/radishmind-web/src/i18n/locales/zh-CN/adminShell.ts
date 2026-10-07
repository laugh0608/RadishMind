export const adminShell = {
  "controlPlane": "S7 · 管理控制台",
  "quotaHeader": "S9 · 配额管理",
  "pricingHeader": "S7 · 价格管理",
  "invitations": "工作区邀请",
  "quota": "应用请求配额",
  "pricing": "模型提供方价格",
  "administration": "管理与路由",
  "invitationHelp": "为此精确工作区创建一次性、有期限的待认领授权意图。仅在已登录的认领人完成原子认领后，成员资格与角色权限才生效。",
  "quotaHelp": "维护精确的开发／测试态应用策略及配额模块提供的 UTC 用量。",
  "pricingHelp": "维护精确、不可变的美元价格修订，供后续开发／测试态请求快照使用。",
  "adminHelp": "审查作用域内的身份信息，再使用明确的开发／测试态配置模块。",
  "tenant": "租户",
  "workspace": "工作区",
  "unavailable": "不可用",
  "authSource": "认证来源",
  "environment": "环境",
  "resourceNavigation": "管理控制台资源",
  "resourcePath": "资源路径",
  "singleOwner": "每次操作一个负责模块",
  "loadingInvitations": "正在加载工作区邀请…",
  "loadingIdentity": "正在加载本地身份管理…",
  "supportingEvidence": "就绪与部署辅助证据",
  "loadingEvidence": "正在加载辅助证据…",
  "boundary": "用户与角色仅消费精确工作区成员目录及服务端内建角色目录。邀请只表达一次性待认领授权意图。邮件发送、全局账户搜索、管理员邀请、自定义角色、生产身份与访问管理及 HTTP 管理员初始化继续关闭。",
  "tasks": {
    "tenant": {
      "label": "租户",
      "scope": "tenant:read"
    },
    "user": {
      "label": "用户",
      "scope": "本地身份"
    },
    "role": {
      "label": "角色",
      "scope": "本地权限"
    },
    "invitations": {
      "label": "邀请",
      "scope": "一次性认领"
    },
    "audit": {
      "label": "审计",
      "scope": "audit:read"
    },
    "provider": {
      "label": "模型提供方",
      "scope": "目录引用"
    },
    "profile": {
      "label": "配置档案",
      "scope": "分配"
    },
    "route": {
      "label": "路由",
      "scope": "代次与版本校验"
    },
    "quota": {
      "label": "配额",
      "scope": "UTC 每日版本校验"
    },
    "pricing": {
      "label": "价格",
      "scope": "美元／百万单位版本校验"
    }
  },
  "status": {
    "authenticated": "已认证",
    "offlineEvidence": "离线证据",
    "offlineWindow": "离线窗口",
    "devControl": "开发／测试态控制",
    "offline": "离线",
    "members": "成员目录",
    "roles": "内建目录",
    "claim": "一次性认领",
    "quota": "UTC 每日版本校验",
    "pricing": "美元／百万单位版本校验",
    "auditCursor": "{{countText}} 条 · 还有后续页",
    "auditPage": "{{countText}} 条 · 已到分页窗口末尾",
    "failure": "读取失败 · {{code}}"
  },
  "auth": {
    "localSession": "本地 Web 会话",
    "offlineFixtures": "离线测试数据",
    "oidcTest": "OIDC 集成测试",
    "signedTest": "签名测试令牌",
    "devHeaders": "开发身份头"
  }
} as const;
