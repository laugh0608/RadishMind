import { uiI18n } from "./instance.ts";
import { workflowInput as zh } from "./locales/zh-CN/workflowInput.ts";
import { workflowInput as en } from "./locales/en-US/workflowInput.ts";

uiI18n.addResourceBundle("zh-CN", "workflow", { input: zh }, true);
uiI18n.addResourceBundle("en-US", "workflow", { input: en }, true);
