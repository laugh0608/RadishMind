export const identityRoles = {
  workspace_reader: { name: "工作区只读成员", summary: "读取工作区应用、运行、工作流、会话、评测与用量元数据。" },
  workspace_builder: { name: "工作区构建者", summary: "创建和执行工作区应用、工作流、会话、配置档案与评测。" },
  workspace_reviewer: { name: "工作区审查者", summary: "构建工作区资源，并执行明确的审查、激活与确认操作。" },
  workspace_admin: { name: "工作区管理员", summary: "管理工作区、破坏性生命周期操作、策略界面、成员与角色分配。" },
  unknown: { name: "未识别的角色", summary: "继续操作前，请审查原始角色键与服务端定义的权限。" },
} as const;
