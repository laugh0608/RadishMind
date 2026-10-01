import { uiI18n } from "./instance.ts";
import { workflowCanvas as zh } from "./locales/zh-CN/workflowCanvas.ts";
import { workflowCanvas as en } from "./locales/en-US/workflowCanvas.ts";

uiI18n.addResourceBundle("zh-CN", "workflow", { canvas: zh }, true);
uiI18n.addResourceBundle("en-US", "workflow", { canvas: en }, true);
