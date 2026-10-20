import "./workflowInspectionProjectionResources.ts";
import { uiI18n } from "./instance.ts";
import { workflowHandoffProjection as zh } from "./locales/zh-CN/workflowHandoffProjection.ts";
import { workflowHandoffProjection as en } from "./locales/en-US/workflowHandoffProjection.ts";

uiI18n.addResourceBundle("zh-CN", "workflow", { projection: zh }, true);
uiI18n.addResourceBundle("en-US", "workflow", { projection: en }, true);
