export const adminShell = {
  "controlPlane": "S7 · 管理控制台",
  "quotaHeader": "S9 · 配额管理",
  "pricingHeader": "S7 · 价格管理",
  "invitations": "工作区邀请",
  "quotaTitle": "应用请求配额",
  "pricingTitle": "模型提供方价格",
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
  },
  "control": "开发 / 测试态管理",
  "offlineOwner": "离线",
  "loadingProvider": "正在加载受控配置…",
  "quotaOwner": "配额 · 开发 / 测试",
  "quotaOwnerTitle": "应用请求配额准入",
  "quotaOwnerHelp": "读取当前 UTC 窗口配额，再审查并确认一次预期版本更新。",
  "loadingQuota": "正在加载应用配额…",
  "pricingOwner": "价格 · 开发 / 测试",
  "pricingOwnerTitle": "不可变模型价格修订",
  "pricingOwnerHelp": "读取精确模型提供方 / 配置档案 / 模型的价格策略，再审查并确认用于后续请求的 CAS 修订。",
  "loadingPricing": "正在加载模型价格…",
  "loadingRead": "正在加载只读记录…",
  "developmentOwner": "{{surface}} · 开发 / 测试",
  "owners": {
    "provider": {
      "title": "模型提供方目录边界",
      "help": "引用已有运行目录，不复制端点或凭据材料。"
    },
    "profile": {
      "title": "模型提供方配置档案分配",
      "help": "在同一开发 / 测试态配置中绑定稳定配置档案引用及能力。"
    },
    "route": {
      "title": "版本化模型路由",
      "help": "先审查不可变候选，再显式切换代次，使后续网关请求使用新配置。"
    }
  }
} as const;
