import { uiI18n } from "./instance.ts";
import { workflowHandoff as zh } from "./locales/zh-CN/workflowHandoff.ts";
import { workflowHandoff as en } from "./locales/en-US/workflowHandoff.ts";

uiI18n.addResourceBundle("zh-CN", "workflow", { handoff: zh }, true);
uiI18n.addResourceBundle("en-US", "workflow", { handoff: en }, true);
