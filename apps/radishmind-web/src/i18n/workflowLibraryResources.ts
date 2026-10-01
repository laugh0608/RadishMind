import { uiI18n } from "./instance.ts";
import { workflowLibrary as zh } from "./locales/zh-CN/workflowLibrary.ts";
import { workflowLibrary as en } from "./locales/en-US/workflowLibrary.ts";

uiI18n.addResourceBundle("zh-CN", "workflow", { library: zh }, true);
uiI18n.addResourceBundle("en-US", "workflow", { library: en }, true);
