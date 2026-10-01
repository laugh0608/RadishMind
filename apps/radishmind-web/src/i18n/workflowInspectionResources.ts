import { uiI18n } from "./instance.ts";
import { workflowInspection as zh } from "./locales/zh-CN/workflowInspection.ts";
import { workflowInspection as en } from "./locales/en-US/workflowInspection.ts";

uiI18n.addResourceBundle("zh-CN", "workflow", { inspection: zh }, true);
uiI18n.addResourceBundle("en-US", "workflow", { inspection: en }, true);
