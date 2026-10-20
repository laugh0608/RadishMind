import { uiI18n } from "./instance.ts";
import { workflowRAGApplication as zh } from "./locales/zh-CN/workflowRAGApplication.ts";
import { workflowRAGApplication as en } from "./locales/en-US/workflowRAGApplication.ts";

uiI18n.addResourceBundle("zh-CN", "workflow", { ragApplication: zh }, true);
uiI18n.addResourceBundle("en-US", "workflow", { ragApplication: en }, true);
