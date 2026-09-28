import { uiI18n } from "./instance.ts";
import { workflowRevision as zh } from "./locales/zh-CN/workflowRevision.ts";
import { workflowRevision as en } from "./locales/en-US/workflowRevision.ts";

uiI18n.addResourceBundle("zh-CN", "workflow", { revision: zh }, true);
uiI18n.addResourceBundle("en-US", "workflow", { revision: en }, true);
