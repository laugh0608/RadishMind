import { uiI18n } from "./instance.ts";
import { workflowHistory as zh } from "./locales/zh-CN/workflowHistory.ts";
import { workflowHistory as en } from "./locales/en-US/workflowHistory.ts";

uiI18n.addResourceBundle("zh-CN", "workflow", { history: zh }, true);
uiI18n.addResourceBundle("en-US", "workflow", { history: en }, true);
