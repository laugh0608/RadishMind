import { uiI18n } from "./instance.ts";
import { workflowRAGExecution as zh } from "./locales/zh-CN/workflowRAGExecution.ts";
import { workflowRAGExecution as en } from "./locales/en-US/workflowRAGExecution.ts";

uiI18n.addResourceBundle("zh-CN", "workflow", { ragExecution: zh }, true);
uiI18n.addResourceBundle("en-US", "workflow", { ragExecution: en }, true);
