import "./workflowWorkspaceProjectionResources.ts";
import "./workflowReadProjectionResources.ts";
import { uiI18n } from "./instance.ts";
import { readEvidence as zh } from "./locales/zh-CN/readEvidence.ts";
import { readEvidence as en } from "./locales/en-US/readEvidence.ts";
uiI18n.addResourceBundle("zh-CN", "workflow", { readEvidence: zh }, true);
uiI18n.addResourceBundle("en-US", "workflow", { readEvidence: en }, true);
