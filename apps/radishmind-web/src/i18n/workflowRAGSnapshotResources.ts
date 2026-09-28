import { uiI18n } from "./instance.ts";
import { workflowRAGSnapshot as zh } from "./locales/zh-CN/workflowRAGSnapshot.ts";
import { workflowRAGSnapshot as en } from "./locales/en-US/workflowRAGSnapshot.ts";

uiI18n.addResourceBundle("zh-CN", "workflow", { ragSnapshot: zh }, true);
uiI18n.addResourceBundle("en-US", "workflow", { ragSnapshot: en }, true);
