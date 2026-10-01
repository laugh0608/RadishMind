import { uiI18n } from "./instance.ts";
import { workflowExecutor as zh } from "./locales/zh-CN/workflowExecutor.ts";
import { workflowExecutor as en } from "./locales/en-US/workflowExecutor.ts";

uiI18n.addResourceBundle("zh-CN", "workflow", { executor: zh }, true);
uiI18n.addResourceBundle("en-US", "workflow", { executor: en }, true);
