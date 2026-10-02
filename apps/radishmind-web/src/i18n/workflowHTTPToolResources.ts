import { uiI18n } from "./instance.ts";
import { workflowHTTPTool as zh } from "./locales/zh-CN/workflowHTTPTool.ts";
import { workflowHTTPTool as en } from "./locales/en-US/workflowHTTPTool.ts";

uiI18n.addResourceBundle("zh-CN", "workflow", { httpTool: zh }, true);
uiI18n.addResourceBundle("en-US", "workflow", { httpTool: en }, true);
