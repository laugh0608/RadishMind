import { uiI18n } from "./instance.ts";
import { workflowReadProjection as zh } from "./locales/zh-CN/workflowReadProjection.ts";
import { workflowReadProjection as en } from "./locales/en-US/workflowReadProjection.ts";
uiI18n.addResourceBundle("zh-CN", "workflow", { projection: zh }, true);
uiI18n.addResourceBundle("en-US", "workflow", { projection: en }, true);
