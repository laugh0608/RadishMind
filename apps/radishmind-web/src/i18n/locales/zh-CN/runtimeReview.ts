export const runtimeReview = {
  "eyebrow": "应用运行审查",
  "title": "运行、审查与观测",
  "intro": "从单次受控请求进入其脱敏记录，再审查当前应用窗口；不推断各记录之间的关联。",
  "application": "应用",
  "workspace": "工作区",
  "lifecycle": "生命周期",
  "tasks": "应用运行审查任务",
  "path": "运行路径",
  "oneTask": "逐项完成任务",
  "boundary": "请求输入与输出仅临时保留。历史记录已脱敏。运营仅并列展示各来源的当前窗口。保存的结果是显式创建、限定于对应负责模块作用域的产物。",
  "loading": "正在加载当前运行功能…",
  "archived": "已归档 · 只读",
  "unavailable": "不可用",
  "tasksBySurface": {
    "run": {
      "label": "运行请求",
      "summary": "临时输入与输出"
    },
    "request": {
      "label": "审查请求",
      "summary": "精确脱敏记录"
    },
    "evidence": {
      "label": "应用证据",
      "summary": "当前加载窗口"
    },
    "results": {
      "label": "已保存结果",
      "summary": "跨会话精确产物"
    }
  }
} as const;
