import { uiI18n } from "./instance.ts";
import { workspacePanel as zh } from "./locales/zh-CN/workspacePanel.ts";
import { workspacePanel as en } from "./locales/en-US/workspacePanel.ts";

uiI18n.addResourceBundle("zh-CN", "applications", { workspacePanel: zh }, true);
uiI18n.addResourceBundle("en-US", "applications", { workspacePanel: en }, true);
