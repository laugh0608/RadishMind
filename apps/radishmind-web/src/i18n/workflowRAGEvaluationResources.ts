import { uiI18n } from "./instance.ts";
import { workflowRAGEvaluation as en } from "./locales/en-US/workflowRAGEvaluation.ts";
import { workflowRAGEvaluation as zh } from "./locales/zh-CN/workflowRAGEvaluation.ts";
uiI18n.addResourceBundle("en-US", "workflow", { ragEvaluation: en }, true);
uiI18n.addResourceBundle("zh-CN", "workflow", { ragEvaluation: zh }, true);
