import { uiI18n } from "./instance.ts";
import { workflowTemplate as zh } from "./locales/zh-CN/workflowTemplate.ts";
import { workflowTemplate as en } from "./locales/en-US/workflowTemplate.ts";

uiI18n.addResourceBundle("zh-CN", "workflow", { template: zh }, true);
uiI18n.addResourceBundle("en-US", "workflow", { template: en }, true);
