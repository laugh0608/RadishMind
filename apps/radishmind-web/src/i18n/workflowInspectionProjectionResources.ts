import { uiI18n } from "./instance.ts";
import { workflowInspectionProjection as zh } from "./locales/zh-CN/workflowInspectionProjection.ts";
import { workflowInspectionProjection as en } from "./locales/en-US/workflowInspectionProjection.ts";

uiI18n.addResourceBundle("zh-CN", "workflow", { projection: zh }, true);
uiI18n.addResourceBundle("en-US", "workflow", { projection: en }, true);
