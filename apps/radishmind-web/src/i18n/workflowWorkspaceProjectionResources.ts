import "./workflowHandoffProjectionResources.ts";
import { uiI18n } from "./instance.ts";
import { workflowWorkspaceProjection as en } from "./locales/en-US/workflowWorkspaceProjection.ts";
import { workflowWorkspaceProjection as zh } from "./locales/zh-CN/workflowWorkspaceProjection.ts";
uiI18n.addResourceBundle("en-US", "workflow", { projection: en }, true);
uiI18n.addResourceBundle("zh-CN", "workflow", { projection: zh }, true);
